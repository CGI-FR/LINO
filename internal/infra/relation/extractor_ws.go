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

package relation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cgi-fr/lino/pkg/relation"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type action string

const (
	ExtractTables action = "extract_relations"
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

// Extract relation from the database.
func (e *WSExtractor) Extract() ([]relation.Relation, *relation.Error) {
	if err := e.Dial(); err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	defer e.Close()

	payload, err := json.Marshal(map[string]string{"schema": e.schema})
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}
	command := CommandMessage{Action: ExtractTables, Payload: payload}

	if err := e.SendMessage(command); err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	result, err := e.ReadResult()
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	if result.Error != "" {
		return nil, &relation.Error{Description: result.Error}
	}

	relations := []relation.Relation{}

	if err := json.Unmarshal(result.Payload, &relations); err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	return relations, nil
}

func (e *WSExtractor) Dial() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var err error
	e.conn, _, err = websocket.Dial(ctx, e.url, &websocket.DialOptions{
		Subprotocols: []string{"lino"},
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

// NewWSExtractorFactory creates a new WS extractor factory.
func NewWSExtractorFactory() *WSExtractorFactory {
	return &WSExtractorFactory{}
}

// WSExtractorFactory exposes methods to create new WS extractors.
type WSExtractorFactory struct{}

// New return a WS extractor
func (e *WSExtractorFactory) New(url string, schema string) relation.Extractor {
	return NewWSExtractor(url, schema)
}
