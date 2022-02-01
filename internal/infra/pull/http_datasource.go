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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cgi-fr/lino/pkg/pull"
	"github.com/rs/zerolog/log"
)

// HTTPDataSourceFactory exposes methods to create new HTTP pullers.
type HTTPDataSourceFactory struct{}

// NewHTTPDataSourceFactory creates a new HTTP datasource factory.
func NewHTTPDataSourceFactory() *HTTPDataSourceFactory {
	return &HTTPDataSourceFactory{}
}

// New return a HTTP puller
func (e *HTTPDataSourceFactory) New(url string, schema string) pull.DataSource {
	return &HTTPDataSource{
		url:    url,
		schema: schema,
	}
}

// HTTPDataSource to read in the pull process.
type HTTPDataSource struct {
	url    string
	schema string
	result io.ReadCloser
}

// Open a connection to the HTTP DB
func (ds *HTTPDataSource) Open() error {
	return nil
}

func (ds *HTTPDataSource) Read(source pull.Table, filter pull.Filter) (pull.RowSet, error) {
	reader, err := ds.RowReader(source, filter)
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

// RowReader iterate over rows in table with filter
func (ds *HTTPDataSource) RowReader(source pull.Table, filter pull.Filter) (pull.RowReader, error) {
	b, err := json.Marshal(struct {
		Values   pull.Row `json:"values"`
		Limit    uint     `json:"limit"`
		Where    string   `json:"where"`
		Distinct bool     `json:"distinct"`
	}{
		Values:   filter.Values,
		Limit:    filter.Limit,
		Where:    filter.Where,
		Distinct: filter.Distinct,
	})
	if err != nil {
		return nil, err
	}
	reqbody := strings.NewReader(string(b))

	url := ds.url + "/data/" + string(source.Name)
	if len(ds.schema) > 0 {
		url = url + "?schema=" + ds.schema
	}

	log.Debug().RawJSON("body", b).Str("url", url).Msg("External connector request")

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, reqbody)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	if pcols := source.Columns; pcols != nil && len(pcols) > 0 {
		pcolsList := []string{}
		for idx := int(0); idx < len(pcols); idx++ {
			pcolsList = append(pcolsList, pcols[idx].Name)
		}
		b, err = json.Marshal(pcolsList)
		if err != nil {
			return nil, err
		}
		pcolsJSON := string(b)
		req.Header.Add("Select-Columns", pcolsJSON)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	ds.result = resp.Body

	return NewJSONRowReader(resp.Body), nil
}

// Close a connection to the HTTP DB
func (ds *HTTPDataSource) Close() error {
	if ds.result != nil {
		ds.result.Close()
	}
	return nil
}
