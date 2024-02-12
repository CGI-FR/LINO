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

package table

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cgi-fr/lino/pkg/table"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type action string

const (
	ExtractTables action = "extract_tables"
)

type CommandMessage struct {
	Id      string          `json:"id"`
	Action  action          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

type ResultMessage struct {
	Id      string          `json:"id"`
	Error   string          `json:"error"`
	Next    bool            `json:"next"`
	Payload json.RawMessage `json:"payload"`
}

// WSExtractor provides table extraction logic from an WS Rest Endpoint.
type WSExtractor struct {
	url      string
	schema   string
	conn     *websocket.Conn
	sequence int
}

// NewWSExtractor creates a new WS extractor.
func NewWSExtractor(url string, schema string) *WSExtractor {
	return &WSExtractor{
		url:    url,
		schema: schema,
	}
}

// Extract tables from the database.
func (e *WSExtractor) Extract(onlyTables bool) ([]table.Table, *table.Error) {
	if err := e.Dial(); err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	defer e.Close()

	payload, err := json.Marshal(map[string]string{"schema": e.schema})
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	command := CommandMessage{Action: ExtractTables, Payload: payload}

	if err := e.SendMessage(command); err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	result, err := e.ReadResult()
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	if result.Error != "" {
		return nil, &table.Error{Description: result.Error}
	}

	tables := []table.Table{}

	if err := json.Unmarshal(result.Payload, &tables); err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	return tables, nil
}

func (e *WSExtractor) Dial() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	u, err := url.Parse(e.url)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	handShakeHeaders := http.Header{}
	if password, ok := u.User.Password(); ok {
		auth := u.User.Username() + ":" + password
		authbase64 := base64.StdEncoding.EncodeToString([]byte(auth))
		handShakeHeaders.Add("Authorization", "Basic "+authbase64)
	}

	e.conn, _, err = websocket.Dial(ctx, e.url, &websocket.DialOptions{
		Subprotocols: []string{"lino"},
		HTTPHeader:   handShakeHeaders,
	})
	return err
}

func (e *WSExtractor) SendMessage(msg CommandMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	defer cancel()
	msg.Id = fmt.Sprintf("%d", e.sequence)
	e.sequence++
	return wsjson.Write(ctx, e.conn, msg)
}

func (e *WSExtractor) ReadResult() (ResultMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	defer cancel()

	result := ResultMessage{}
	err := wsjson.Read(ctx, e.conn, &result)
	return result, err
}

func (e *WSExtractor) Close() {
	if e.conn != nil {
		e.conn.Close(websocket.StatusNormalClosure, "")
	}
}

func (e *WSExtractor) Count(tableName string) (int, *table.Error) {
	return 0, nil // TODO
}

// NewWSExtractorFactory creates a new WS extractor factory.
func NewWSExtractorFactory() *WSExtractorFactory {
	return &WSExtractorFactory{}
}

// WSExtractorFactory exposes methods to create new WS extractors.
type WSExtractorFactory struct{}

// New return a WS extractor
func (e *WSExtractorFactory) New(url string, schema string) table.Extractor {
	return NewWSExtractor(url, schema)
}
