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

package push

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cgi-fr/lino/pkg/push"
	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Action string

//  "pull_open", "push_open", "push_data", "push_commit", "push_close"
const (
	PushOpen   Action = "push_open"
	PushData   Action = "push_data"
	PushCommit Action = "push_commit"
	PushClose  Action = "push_close"
)

type CommandMessage struct {
	Id      string          `json:"id"`
	Action  Action          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

type WritePayload struct {
	Table      string   `json:"table"`
	Row        push.Row `json:"row"`
	Conditions push.Row `json:"conditions"`
}

type OpenPayload struct {
	Schema             string   `json:"schema"`
	Tables             []string `json:"tables"`
	Mode               string   `json:"mode"`
	DisableConstraints bool     `json:"disable_constraints"`
}

type ResultMessage struct {
	Id      string          `json:"id"`
	Error   string          `json:"error"`
	Next    bool            `json:"next"`
	Payload json.RawMessage `json:"payload"`
}

// WebSocketDataDestinationFactory exposes methods to create new websocket pusher.
type WebSocketDataDestinationFactory struct{}

// NewWebSocketDataDestinationFactory creates a new websocket datadestination factory.
func NewWebSocketDataDestinationFactory() *WebSocketDataDestinationFactory {
	return &WebSocketDataDestinationFactory{}
}

// New return a web socket pusher
func (e *WebSocketDataDestinationFactory) New(url string, schema string) push.DataDestination {
	return NewWebSocketDataDestination(url, schema)
}

// WebSocketDataDestination write data to a web socket endpoint.
type WebSocketDataDestination struct {
	url                string
	schema             string
	mode               push.Mode
	disableConstraints bool
	conn               *websocket.Conn
	sequence           uint
}

// NewWebSocketDataDestination creates a new web socket datadestination.
func NewWebSocketDataDestination(url string, schema string) *WebSocketDataDestination {
	return &WebSocketDataDestination{
		url:                url,
		schema:             schema,
		mode:               push.Insert,
		disableConstraints: false,
	}
}

func (dd *WebSocketDataDestination) SendMessageAndReadResult(msg CommandMessage) *push.Error {
	ctx := context.Background()

	msg.Id = fmt.Sprintf("%d", dd.sequence)
	dd.sequence++
	log.Trace().RawJSON("payload", msg.Payload).Str("id", msg.Id).Str("action_msg", string(msg.Action)).Msg("send message to server")

	if err := wsjson.Write(ctx, dd.conn, msg); err != nil {
		return &push.Error{Description: err.Error()}
	}

	result := ResultMessage{}
	if err := wsjson.Read(ctx, dd.conn, &result); err != nil {
		return &push.Error{Description: err.Error()}
	}
	log.Trace().RawJSON("payload", result.Payload).Str("id", result.Id).Str("error", result.Error).Msg("receive message from server")
	if result.Id != msg.Id {
		return &push.Error{Description: fmt.Sprintf("server send a response with different ID want=%s, receive=%s", msg.Id, result.Id)}
	}

	if result.Error != "" {
		return &push.Error{Description: string(result.Payload)}
	}

	return nil
}

// Open web socket connection
func (dd *WebSocketDataDestination) Open(plan push.Plan, mode push.Mode, disableConstraints bool) *push.Error {
	log.Debug().Str("url", dd.url).Str("schema", dd.schema).Str("mode", mode.String()).Bool("disableConstraints", disableConstraints).Msg("open web socket destination")
	dd.mode = mode
	dd.disableConstraints = disableConstraints

	payload := OpenPayload{
		Schema:             dd.schema,
		Tables:             []string{},
		Mode:               mode.String(),
		DisableConstraints: disableConstraints,
	}

	for _, table := range plan.Tables() {
		payload.Tables = append(payload.Tables, table.Name())
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	u, err := url.Parse(dd.url)
	if err != nil {
		return &push.Error{Description: fmt.Sprintf("failed to parse url: %s", err.Error())}
	}

	handShakeHeaders := http.Header{}
	if password, ok := u.User.Password(); ok {
		auth := u.User.Username() + ":" + password
		authbase64 := base64.StdEncoding.EncodeToString([]byte(auth))
		handShakeHeaders.Add("Authorization", "Basic "+authbase64)
	}

	dd.conn, _, err = websocket.Dial(ctx, dd.url, &websocket.DialOptions{
		Subprotocols: []string{"lino"},
		HTTPHeader:   handShakeHeaders,
	})
	if err != nil {
		log.Err(err).Str("url", dd.url).Str("schema", dd.schema).Msg("error while dialing connexion")
		return &push.Error{Description: err.Error()}
	}

	msg := CommandMessage{Action: PushOpen, Payload: data}
	if err := dd.SendMessageAndReadResult(msg); err != nil {
		dd.Close()
		return err
	}

	return nil
}

// Close web socket connection
func (dd *WebSocketDataDestination) Close() *push.Error {
	log.Debug().Str("url", dd.url).Str("schema", dd.schema).Msg("close web socket destination")
	defer dd.conn.Close(websocket.StatusNormalClosure, "")
	msg := CommandMessage{Action: PushClose, Payload: json.RawMessage([]byte("{}"))}
	if err := dd.SendMessageAndReadResult(msg); err != nil {
		return err
	}

	return nil
}

// Commit web socket for connection
func (dd *WebSocketDataDestination) Commit() *push.Error {
	log.Debug().Str("url", dd.url).Str("schema", dd.schema).Msg("commit web socket destination")
	msg := CommandMessage{Action: PushCommit}

	if err := dd.SendMessageAndReadResult(msg); err != nil {
		return err
	}

	return nil
}

// RowWriter return web socket table writer
func (dd *WebSocketDataDestination) RowWriter(table push.Table) (push.RowWriter, *push.Error) {
	return &WebSocketRowWriter{dd, table}, nil
}

type WebSocketRowWriter struct {
	dd    *WebSocketDataDestination
	table push.Table
}

func (rw *WebSocketRowWriter) Write(row push.Row, translatedKeys push.Cache) *push.Error {
	conditions := push.Row{}
	if rw.dd.mode == push.Update || rw.dd.mode == push.Delete {
		for _, pk := range rw.table.PrimaryKey() {
			var find bool
			conditions[pk], find = row[pk]
			if !find {
				return &push.Error{Description: fmt.Sprintf("Expected primary key %s in row %v", pk, row)}
			}
			delete(row, pk)
		}
	}

	payload := WritePayload{
		Table:      rw.table.Name(),
		Row:        row,
		Conditions: conditions,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	msg := CommandMessage{Action: PushData, Payload: data}

	if err := rw.dd.SendMessageAndReadResult(msg); err != nil {
		return err
	}

	return nil
}
