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

package dataconnector

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cgi-fr/lino/pkg/dataconnector"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type action string

const (
	ExtractTables action = "ping"
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

type WSDataPingerFactory struct{}

// NewWSDataPinger creates a new HTTP pinger.
func NewWSDataPingerFactory() *WSDataPingerFactory {
	return &WSDataPingerFactory{}
}

func (pdpf WSDataPingerFactory) New(url string) dataconnector.DataPinger {
	return NewWSDataPinger(url)
}

func NewWSDataPinger(url string) WSDataPinger {
	return WSDataPinger{url: url}
}

type WSDataPinger struct {
	url      string
	conn     *websocket.Conn
	sequence int
}

func (e WSDataPinger) Ping() *dataconnector.Error {
	if err := e.Dial(); err != nil {
		return &dataconnector.Error{Description: err.Error()}
	}

	defer e.Close()

	command := CommandMessage{Action: ExtractTables}

	if err := e.SendMessage(command); err != nil {
		return &dataconnector.Error{Description: err.Error()}
	}

	result, err := e.ReadResult()
	if err != nil {
		return &dataconnector.Error{Description: err.Error()}
	}

	if result.Error != "" {
		return &dataconnector.Error{Description: "Error from server : " + string(result.Payload)}
	}

	return nil
}

func (e *WSDataPinger) Dial() error {
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

func (e *WSDataPinger) SendMessage(msg CommandMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	defer cancel()
	msg.Id = fmt.Sprintf("%d", e.sequence)
	e.sequence++
	return wsjson.Write(ctx, e.conn, msg)
}

func (e *WSDataPinger) ReadResult() (ResultMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	defer cancel()

	result := ResultMessage{}
	err := wsjson.Read(ctx, e.conn, &result)
	return result, err
}

func (e *WSDataPinger) Close() {
	if e.conn != nil {
		e.conn.Close(websocket.StatusNormalClosure, "")
	}
}
