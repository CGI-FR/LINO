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
	"encoding/json"
	"io"
	"net/http"

	"github.com/cgi-fr/lino/pkg/relation"
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
func (e *HTTPExtractor) Extract() ([]relation.Relation, *relation.Error) {
	resp, err := http.Get("http://httpmock:1080/api/v1/relations")
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}
	container := struct {
		Version   string
		Relations []relation.Relation
	}{
		"",
		[]relation.Relation{},
	}
	err = json.Unmarshal(body, &container)
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}
	return container.Relations, nil
}

// NewHTTPExtractorFactory creates a new HTTP extractor factory.
func NewHTTPExtractorFactory() *HTTPExtractorFactory {
	return &HTTPExtractorFactory{}
}

// HTTPExtractorFactory exposes methods to create new HTTP extractors.
type HTTPExtractorFactory struct{}

// New return a HTTP extractor
func (e *HTTPExtractorFactory) New(url string, schema string) relation.Extractor {
	return NewHTTPExtractor(url, schema)
}
