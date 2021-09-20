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
	"net/http"

	"github.com/cgi-fr/lino/pkg/dataconnector"
)

type HTTPDataPingerFactory struct{}

// NewHTTPDataPinger creates a new HTTP pinger.
func NewHTTPDataPingerFactory() *HTTPDataPingerFactory {
	return &HTTPDataPingerFactory{}
}

func (pdpf HTTPDataPingerFactory) New(url string) dataconnector.DataPinger {
	return NewHTTPDataPinger(url)
}

func NewHTTPDataPinger(url string) HTTPDataPinger {
	return HTTPDataPinger{url}
}

type HTTPDataPinger struct {
	url string
}

func (pdp HTTPDataPinger) Ping() *dataconnector.Error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodHead, pdp.url, nil)
	if err != nil {
		return &dataconnector.Error{Description: err.Error()}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &dataconnector.Error{Description: err.Error()}
	}
	resp.Body.Close()
	return nil
}
