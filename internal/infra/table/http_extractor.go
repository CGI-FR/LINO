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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cgi-fr/lino/pkg/table"
	"github.com/rs/zerolog/log"
)

// HTTPExtractor provides table extraction logic from an HTTP Rest Endpoint.
type HTTPExtractor struct {
	url    string
	schema string
}

// NewHTTPExtractor creates a new HTTP extractor.
func NewHTTPExtractor(url string, schema string) *HTTPExtractor {
	return &HTTPExtractor{
		url:    url,
		schema: schema,
	}
}

// Extract tables from the database.
func (e *HTTPExtractor) Extract(onlyTables bool, withDBInfos bool) ([]table.Table, *table.Error) {
	url := e.url + "/tables"
	if len(e.schema) > 0 {
		url = url + "?schema=" + e.schema
	}

	log.Debug().Str("url", url).Msg("External connector request")

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	container := struct {
		Version string
		Tables  []table.Table
	}{
		"",
		[]table.Table{},
	}
	err = json.Unmarshal(body, &container)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	return container.Tables, nil
}

func (e *HTTPExtractor) Count(tableName string) (int, *table.Error) {
	url := e.url + "/data/" + tableName + "/count"

	if len(e.schema) > 0 {
		url = url + "?schema=" + e.schema
	}

	log.Debug().Str("url", url).Msg("External connector request")

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return 0, &table.Error{Description: err.Error()}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, &table.Error{Description: err.Error()}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, &table.Error{Description: fmt.Sprintf("HTTP request returned status code %d", resp.StatusCode)}
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, &table.Error{Description: err.Error()}
	}

	container := struct {
		Version string
		Count   int
	}{
		"",
		0,
	}
	err = json.Unmarshal(body, &container)
	if err != nil {
		return 0, &table.Error{Description: err.Error()}
	}
	return container.Count, nil
}

// NewHTTPExtractorFactory creates a new HTTP extractor factory.
func NewHTTPExtractorFactory() *HTTPExtractorFactory {
	return &HTTPExtractorFactory{}
}

// HTTPExtractorFactory exposes methods to create new HTTP extractors.
type HTTPExtractorFactory struct{}

// New return a HTTP extractor
func (e *HTTPExtractorFactory) New(url string, schema string) table.Extractor {
	return NewHTTPExtractor(url, schema)
}
