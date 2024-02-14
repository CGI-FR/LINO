// Copyright (C) 2021 CGI France
//
// This file is part of LINO.
//
// LINO is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// LINO is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with LINO.  If not, see <http://www.gnu.org/licenses/>.

package pull

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/cgi-fr/lino/pkg/pull"
	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type action string

// "pull_open", "push_open", "push_data", "push_commit", "push_close"
const (
	PullOpen action = "pull_open"
)

type CommandMessage struct {
	Id      string          `json:"id"`
	Action  action          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

type PullPayload struct {
	Schema  string      `json:"schema"`
	Table   string      `json:"table"`
	Columns []string    `json:"columns"`
	Filter  pull.Filter `json:"filter"`
}

type ResultMessage struct {
	Id      string          `json:"id"`
	Error   string          `json:"error"`
	Next    bool            `json:"next"`
	Payload json.RawMessage `json:"payload"`
}

// WSDataSourceFactory exposes methods to create new WS pullers.
type WSDataSourceFactory struct{}

// NewWSDataSourceFactory creates a new WS datasource factory.
func NewWSDataSourceFactory() *WSDataSourceFactory {
	return &WSDataSourceFactory{}
}

// New return a WS puller
func (e *WSDataSourceFactory) New(url string, schema string) pull.DataSource {
	return &WSDataSource{
		url:    url,
		schema: schema,
	}
}

// WSDataSource to read in the pull process.
type WSDataSource struct {
	sync.Mutex
	url      string
	schema   string
	conn     *websocket.Conn
	sequence uint
	cancels  []context.CancelFunc
	results  map[string]chan ResultMessage
	stop     chan bool
	localDS  *WSDataSource
}

// Open a connection to the WS DB
func (ds *WSDataSource) Open() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	u, err := url.Parse(ds.url)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	handShakeHeaders := http.Header{}
	if password, ok := u.User.Password(); ok {
		auth := u.User.Username() + ":" + password
		authbase64 := base64.StdEncoding.EncodeToString([]byte(auth))
		handShakeHeaders.Add("Authorization", "Basic "+authbase64)
	}

	ds.conn, _, err = websocket.Dial(ctx, ds.url, &websocket.DialOptions{
		Subprotocols: []string{"lino"},
		HTTPHeader:   handShakeHeaders,
	})
	log.Trace().Msg("connected to ws")

	ds.results = map[string]chan ResultMessage{}
	ds.stop = make(chan bool)

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		ds.cancels = append(ds.cancels, cancel)

		for {
			result := ResultMessage{}
			err := wsjson.Read(ctx, ds.conn, &result)
			if err != nil {
				log.Info().AnErr("error", err).Msg("Error reading result stop consuming messages")

				break
			}
			log.Trace().Str("Error", result.Error).Str("id", result.Id).Str("payload", string(result.Payload)).Msg("receive message")

			ds.Lock()
			resultChan := ds.results[result.Id]
			resultChan <- result
			if !result.Next {
				close(resultChan)
				delete(ds.results, result.Id)
				ds.Unlock()

				continue
			}
			ds.Unlock()
		}
	}()

	return err
}

func (ds *WSDataSource) Read(source pull.Table, filter pull.Filter) (pull.RowSet, error) {
	reader, err := ds.SharedRowReader(source, filter)
	if err != nil {
		return nil, err
	}

	result := pull.RowSet{}
	for reader.Next() {
		result = append(result, reader.Value())
	}

	if reader.Error() != nil {
		return result, fmt.Errorf("%w", reader.Error())
	}

	return result, nil
}

// RowReader iterate over rows in table with filter using a shared WS
func (ds *WSDataSource) SharedRowReader(source pull.Table, filter pull.Filter) (pull.RowReader, error) {
	pcolsList := []string{}
	if pcols := source.Columns; len(pcols) > 0 && source.ExportMode != pull.ExportModeAll {
		for idx := int(0); idx < len(pcols); idx++ {
			pcolsList = append(pcolsList, pcols[idx].Name)
		}
	}

	payload, err := json.Marshal(PullPayload{ds.schema, string(source.Name), pcolsList, filter})
	if err != nil {
		return nil, err
	}
	command := CommandMessage{Action: PullOpen, Payload: payload}

	result, err := ds.SendMessageAndReadResultStream(command)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// RowReader iterate over rows in table with filter using a dedicated ws
func (ds *WSDataSource) RowReader(source pull.Table, filter pull.Filter) (pull.RowReader, error) {
	if ds.localDS != nil {
		return ds.localDS.SharedRowReader(source, filter)
	}

	log.Trace().Str("table", string(source.Name)).Msg("new ws for read table")

	ds.localDS = &WSDataSource{url: ds.url, schema: ds.schema}

	go func() {
		<-ds.stop
		ds.localDS.Close()
	}()

	if err := ds.localDS.Open(); err != nil {
		return nil, err
	}

	return ds.localDS.SharedRowReader(source, filter)
}

func (ds *WSDataSource) SendMessageAndReadResultStream(msg CommandMessage) (ResultStream, error) {
	ctx, cancel := context.WithCancel(context.Background())
	ds.cancels = append(ds.cancels, cancel)

	resultChan := make(chan ResultMessage, 1)
	stream := ResultStream{payloadStream: resultChan}

	ds.Lock()
	msg.Id = fmt.Sprintf("%d", ds.sequence)
	ds.results[msg.Id] = resultChan
	ds.sequence++
	ds.Unlock()

	if err := wsjson.Write(ctx, ds.conn, msg); err != nil {
		return stream, err
	}

	return stream, nil
}

// Close a connection to the WS DB
func (ds *WSDataSource) Close() error {
	close(ds.stop)
	for _, cancel := range ds.cancels {
		cancel()
	}
	if ds.conn != nil {
		log.Trace().Msg("Close WS")

		ds.conn.Close(websocket.StatusNormalClosure, "")
	}

	return nil
}

type ResultStream struct {
	payloadStream chan ResultMessage
	lastMessage   ResultMessage
	lastRow       pull.Row
	err           error
}

// Next return true if Next Value is present
func (rs *ResultStream) Next() bool {
	lastMessage := <-rs.payloadStream

	if lastMessage.Error == "Error" {
		rs.err = fmt.Errorf("Receive error from web socket server : %s ", string(lastMessage.Payload))
		return false
	}
	if lastMessage.Id == "" {
		return false
	}

	if !lastMessage.Next {
		log.Trace().Str("Id", lastMessage.Id).Msg("End of web socket stream")
		return false
	}

	rs.lastRow = pull.Row{}
	if err := json.Unmarshal(lastMessage.Payload, &rs.lastRow); err != nil {
		log.Error().AnErr("error", err).Str("Id", lastMessage.Id).Str("payload", string(lastMessage.Payload)).Msg("Error unmarshal row")
		rs.err = err
		return false
	}
	rs.lastMessage = lastMessage
	return true
}

// Value return the current Row
func (rs *ResultStream) Value() pull.Row {
	if rs.lastMessage.Id != "" {
		return rs.lastRow
	}
	panic("Value is not valid before call Next")
}

func (rs *ResultStream) Error() error {
	return rs.err
}

func (rs *ResultStream) Close() error {
	return nil
}
