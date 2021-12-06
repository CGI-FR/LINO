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
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/cgi-fr/lino/pkg/table"
	"github.com/xo/dburl"
)

// SQLExtractor provides table extraction logic from SQL database.
type SQLExtractor struct {
	url     string
	schema  string
	dialect Dialect
}

type Dialect interface {
	// TablesSQL return SQL command to list tables from meta data
	TablesSQL(schema string) string
	// SequencesSQL return SQL command to list sequences from meta data
	SequencesSQL(schema string) string
	// UpdateSequenceSQL return SQL Command to update sequence to max +1  of tablename/column values
	UpdateSequenceSQL(schema string, sequence string, tableName string, column string) string
}

// NewSQLExtractor creates a new SQL extractor.
func NewSQLExtractor(url string, schema string, dialect Dialect) *SQLExtractor {
	return &SQLExtractor{
		url:     url,
		schema:  schema,
		dialect: dialect,
	}
}

// Extract tables from the database.
func (e *SQLExtractor) Extract() ([]table.Table, *table.Error) {
	db, err := dburl.Open(e.url)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	SQL := e.dialect.TablesSQL(e.schema)

	rows, err := db.Query(SQL)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	tables := map[string]*table.Table{}

	var (
		tableSchema string
		tableName   string
		keyColumns  string
	)

	for rows.Next() {
		err := rows.Scan(&tableSchema, &tableName, &keyColumns)
		if err != nil {
			return nil, &table.Error{Description: err.Error()}
		}

		table := table.Table{

			Name: tableName,
			Keys: strings.Split(keyColumns, ","),
		}
		tables[table.Name] = &table
	}
	err = rows.Err()
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	rows, err = db.Query(e.dialect.SequencesSQL(e.schema))
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	var sequenceName string

	for rows.Next() {
		err := rows.Scan(&sequenceName)
		if err != nil {
			return nil, &table.Error{Description: err.Error()}
		}

		log.Debug().Str("sequence", sequenceName).Msg("find new sequence")

		for _, tbl := range tables {
			sequences := tbl.Sequences
			for _, key := range tbl.Keys {
				if strings.Contains(sequenceName, tbl.Name) && strings.Contains(sequenceName, key) {
					log.Debug().Str("sequence", sequenceName).Str("table", tbl.Name).Str("key", key).Msg("sequence match")
					sequences = append(sequences, table.Sequence{Name: sequenceName, Key: key})
				}
			}
			tbl.Sequences = sequences
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	result := []table.Table{}
	for _, table := range tables {
		result = append(result, *table)
	}
	return result, nil
}

// Update sequence
func (e *SQLExtractor) UpdateSequence(sequence string, tableName string, column string) *table.Error {
	db, err := dburl.Open(e.url)
	if err != nil {
		return &table.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return &table.Error{Description: err.Error()}
	}

	SQL := e.dialect.UpdateSequenceSQL(e.schema, sequence, tableName, column)
	log.Debug().Str("sql", SQL).Msg("SQL to update sequence")
	_, err = db.Query(SQL)
	if err != nil {
		return &table.Error{Description: err.Error()}
	}

	return nil
}
