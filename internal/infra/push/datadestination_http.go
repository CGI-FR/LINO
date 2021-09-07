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
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cgi-fr/lino/pkg/push"
	"github.com/rs/zerolog/log"
)

// HTTPDataDestinationFactory exposes methods to create new HTTP pusher.
type HTTPDataDestinationFactory struct{}

// NewHTTPDataDestinationFactory creates a new HTTP datadestination factory.
func NewHTTPDataDestinationFactory() *HTTPDataDestinationFactory {
	return &HTTPDataDestinationFactory{}
}

// New return a HTTP pusher
func (e *HTTPDataDestinationFactory) New(url string, schema string) push.DataDestination {
	return NewHTTPDataDestination(url, schema)
}

// HTTPDataDestination write data to a HTTP endpoint.
type HTTPDataDestination struct {
	url       string
	schema    string
	rowWriter map[string]*HTTPRowWriter
}

// NewHTTPDataDestination creates a new HTTP datadestination.
func NewHTTPDataDestination(url string, schema string) *HTTPDataDestination {
	return &HTTPDataDestination{
		url:       url,
		schema:    schema,
		rowWriter: map[string]*HTTPRowWriter{},
	}
}

// Open HTTP Connection
func (dd *HTTPDataDestination) Open(plan push.Plan, mode push.Mode, disableConstraints bool) *push.Error {
	log.Debug().Str("url", dd.url).Str("schema", dd.schema).Str("mode", mode.String()).Bool("disableConstraints", disableConstraints).Msg("open HTTP destination")
	return nil
}

// Close HTTP connections
func (dd *HTTPDataDestination) Close() *push.Error {
	log.Debug().Str("url", dd.url).Str("schema", dd.schema).Msg("close HTTP destination")
	for _, rw := range dd.rowWriter {
		err := rw.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Commit HTTP for connection
func (dd *HTTPDataDestination) Commit() *push.Error {
	log.Debug().Str("url", dd.url).Str("schema", dd.schema).Msg("commit HTTP destination")
	return nil
}

// RowWriter return HTTP table writer
func (dd *HTTPDataDestination) RowWriter(table push.Table) (push.RowWriter, *push.Error) {
	rw, ok := dd.rowWriter[table.Name()]
	if ok {
		return rw, nil
	}

	log.Debug().Str("url", dd.url).Str("schema", dd.schema).Str("table", table.Name()).Msg("build row writer HTTP destination")

	url := dd.url + "/data/" + table.Name()
	if len(dd.schema) > 0 {
		url = url + "?schema=" + dd.schema
	}

	pr, pw := io.Pipe()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, ioutil.NopCloser(pr))
	if err != nil {
		return nil, &push.Error{Description: err.Error()}
	}

	rw = NewHTTPRowWriter(table, dd, req, pw)
	dd.rowWriter[table.Name()] = rw

	go rw.Request()

	return rw, nil
}

// HTTPRowWriter write data to a HTTP table.
type HTTPRowWriter struct {
	table push.Table
	dd    *HTTPDataDestination
	req   *http.Request
	buf   io.WriteCloser
}

// NewHTTPRowWriter creates a new HTTP row writer.
func NewHTTPRowWriter(table push.Table, dd *HTTPDataDestination, req *http.Request, buf io.WriteCloser) *HTTPRowWriter {
	return &HTTPRowWriter{
		table: table,
		dd:    dd,
		req:   req,
		buf:   buf,
	}
}

// Request
func (rw *HTTPRowWriter) Request() {
	resp, err := http.DefaultClient.Do(rw.req)
	if err != nil {
		log.Error().Err(err).Str("url", rw.dd.url).Str("schema", rw.dd.schema).Str("table", rw.table.Name()).Str("status", resp.Status).Msg("response")
	}

	log.Debug().Str("url", rw.dd.url).Str("schema", rw.dd.schema).Str("table", rw.table.Name()).Str("status", resp.Status).Msg("response")
}

// Write
func (rw *HTTPRowWriter) Write(row push.Row) *push.Error {
	jsonline, _ := export(row)
	log.Debug().Str("url", rw.dd.url).Str("schema", rw.dd.schema).Str("table", rw.table.Name()).RawJSON("data", jsonline).Msg("write")
	_, err := rw.buf.Write(jsonline)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	_, err = rw.buf.Write([]byte("\n"))
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

// close table writer
func (rw *HTTPRowWriter) Close() *push.Error {
	log.Debug().Str("url", rw.dd.url).Str("schema", rw.dd.schema).Str("table", rw.table.Name()).Msg("close")
	_, err := rw.buf.Write([]byte("closed\n"))
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	rw.buf.Close()
	return nil
}

// Export rows in JSON format.
func export(r push.Row) ([]byte, *push.Error) {
	jsonString, err := json.Marshal(r)
	if err != nil {
		return nil, &push.Error{Description: err.Error()}
	}
	return jsonString, nil
}
