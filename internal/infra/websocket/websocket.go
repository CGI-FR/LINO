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

package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cgi-fr/lino/pkg/table"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type action string

//  "pull_open", "push_open", "push_data", "push_commit", "push_close"
const (
	Ping             action = "ping"
	ExtractTables    action = "extract_tables"
	ExtractRelations action = "extract_relations"
	PullOpen         action = "pull_open"
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

func New(url string) Client {
	return Client{url: url}
}

type Client struct {
	url      string
	conn     *websocket.Conn
	sequence uint
}

func (c *Client) Ping() error {
	if err := c.Dial(); err != nil {
		return err
	}

	defer c.Close()

	if err := c.SendMessage(CommandMessage{Action: Ping}); err != nil {
		return err
	}

	result, err := c.ReadResult()
	if err != nil {
		return err
	}

	if result.Error != "" {
		return fmt.Errorf(result.Error)
	}

	return nil
}

func (c *Client) ExtractTables(schema string) ([]table.Table, error) {
	if err := c.Dial(); err != nil {
		return nil, err
	}

	defer c.Close()

	payload, err := json.Marshal(map[string]string{"shema": schema})
	if err != nil {
		return nil, err
	}
	command := CommandMessage{Action: ExtractTables, Payload: payload}

	if err := c.SendMessage(command); err != nil {
		return nil, err
	}

	result, err := c.ReadResult()
	if err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf(result.Error)
	}

	tables := []table.Table{}

	if err := json.Unmarshal(result.Payload, &tables); err != nil {
		return nil, err
	}

	return tables, nil
}

func (c *Client) SendMessage(msg CommandMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	defer cancel()
	msg.Id = fmt.Sprintf("%d", c.sequence)
	c.sequence++
	return wsjson.Write(ctx, c.conn, msg)
}

func (c *Client) ReadResult() (ResultMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	defer cancel()

	result := ResultMessage{}
	err := wsjson.Read(ctx, c.conn, &result)
	return result, err
}

func (c *Client) Dial() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var err error
	c.conn, _, err = websocket.Dial(ctx, c.url, &websocket.DialOptions{
		Subprotocols: []string{"lino"},
	})
	return err
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close(websocket.StatusNormalClosure, "")
	}
}

type Protocol struct{}
