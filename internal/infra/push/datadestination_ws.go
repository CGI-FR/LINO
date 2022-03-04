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
	"time"

	"github.com/cgi-fr/lino/pkg/push"
	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

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

// Open web socket connection
func (dd *WebSocketDataDestination) Open(plan push.Plan, mode push.Mode, disableConstraints bool) *push.Error {
	log.Debug().Str("url", dd.url).Str("schema", dd.schema).Str("mode", mode.String()).Bool("disableConstraints", disableConstraints).Msg("open web socket destination")
	dd.mode = mode
	dd.disableConstraints = disableConstraints

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var err error
	dd.conn, _, err = websocket.Dial(ctx, dd.url, nil)
	if err != nil {
		log.Err(err).Str("url", dd.url).Str("schema", dd.schema).Msg("error while dialing connexion")
		return &push.Error{Description: err.Error()}
	}

	return nil
}

// Close web socket connection
func (dd *WebSocketDataDestination) Close() *push.Error {
	log.Debug().Str("url", dd.url).Str("schema", dd.schema).Msg("close web socket destination")
	dd.conn.Close(websocket.StatusNormalClosure, "")
	return nil
}

// Commit web socket for connection
func (dd *WebSocketDataDestination) Commit() *push.Error {
	log.Debug().Str("url", dd.url).Str("schema", dd.schema).Msg("commit web socket destination")

	if err := wsjson.Write(context.Background(), dd.conn, "commit"); err != nil {
		log.Err(err).Str("url", dd.url).Str("schema", dd.schema).Msg("error while sending commit")
		return &push.Error{Description: err.Error()}
	}

	return nil
}

// RowWriter return web socket table writer
func (dd *WebSocketDataDestination) RowWriter(table push.Table) (push.RowWriter, *push.Error) {
	return dd, nil
}

func (dd *WebSocketDataDestination) Write(row push.Row) *push.Error {
	log.Debug().Str("url", dd.url).Str("schema", dd.schema).Interface("data", row).Msg("write to web socket destination")

	if err := wsjson.Write(context.Background(), dd.conn, row); err != nil {
		log.Err(err).Str("url", dd.url).Str("schema", dd.schema).Msg("error while sending data")
		return &push.Error{Description: err.Error()}
	}
	if err := wsjson.Write(context.Background(), dd.conn, "test"); err != nil {
		log.Err(err).Str("url", dd.url).Str("schema", dd.schema).Msg("error while sending commit")
		return &push.Error{Description: err.Error()}
	}

	log.Info().Str("url", dd.url).Str("schema", dd.schema).Interface("data", row).Msg("sent data")

	return nil
}
