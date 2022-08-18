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
	"github.com/cgi-fr/lino/internal/infra/websocket"
	"github.com/cgi-fr/lino/pkg/dataconnector"
)

type WSDataPingerFactory struct{}

// NewWSDataPinger creates a new HTTP pinger.
func NewWSDataPingerFactory() *WSDataPingerFactory {
	return &WSDataPingerFactory{}
}

func (pdpf WSDataPingerFactory) New(url string) dataconnector.DataPinger {
	return NewWSDataPinger(url)
}

func NewWSDataPinger(url string) WSDataPinger {
	return WSDataPinger{url}
}

type WSDataPinger struct {
	url string
}

func (pdp WSDataPinger) Ping() *dataconnector.Error {
	client := websocket.New(pdp.url)

	if err := client.Ping(); err != nil {
		return &dataconnector.Error{Description: err.Error()}
	}
	return nil
}
