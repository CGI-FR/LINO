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
	"github.com/cgi-fr/lino/internal/infra/websocket"
	"github.com/cgi-fr/lino/pkg/relation"
)

// WSExtractor provides table extraction logic from an WS Rest Endpoint.
type WSExtractor struct {
	url    string
	schema string
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
	client := websocket.New(e.url)

	relations, err := client.ExtractRelations(e.schema)
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	return relations, nil
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
