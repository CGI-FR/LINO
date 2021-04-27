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
	"github.com/cgi-fr/lino/pkg/relation"
	"github.com/xo/dburl"
)

// SQLExtractor provides relation extraction logic from SQL database.
type SQLExtractor struct {
	url     string
	schema  string
	dialect Dialect
}

type Dialect interface {
	SQL(schema string) string
}

// NewSQLExtractor creates a new SQL extractor.
func NewSQLExtractor(url string, schema string, dialect Dialect) *SQLExtractor {
	return &SQLExtractor{
		url:     url,
		schema:  schema,
		dialect: dialect,
	}
}

// Extract relations from the database.
func (e *SQLExtractor) Extract() ([]relation.Relation, *relation.Error) {
	db, err := dburl.Open(e.url)

	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	rows, err := db.Query(e.dialect.SQL(e.schema))
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	relations := []relation.Relation{}

	var (
		relationName string
		sourceTable  string
		sourceColumn string
		targetTable  string
		targetColumn string
	)

	for rows.Next() {
		err := rows.Scan(&relationName, &sourceTable, &sourceColumn, &targetTable, &targetColumn)
		if err != nil {
			return nil, &relation.Error{Description: err.Error()}
		}

		relation := relation.Relation{
			Name: relationName,
			Parent: relation.Table{
				Name: targetTable,
				Keys: []string{targetColumn},
			},
			Child: relation.Table{
				Name: sourceTable,
				Keys: []string{sourceColumn},
			},
		}
		relations = append(relations, relation)
	}
	err = rows.Err()
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	return relations, nil
}
