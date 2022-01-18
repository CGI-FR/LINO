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

package sequence

import (
	"github.com/rs/zerolog/log"

	"github.com/cgi-fr/lino/pkg/sequence"
	"github.com/xo/dburl"
)

// SQLUpdator provides table extraction logic from SQL database.
type SQLUpdator struct {
	url     string
	schema  string
	dialect Dialect
}

type Dialect interface {
	// SequencesSQL return SQL command to list sequences from meta data
	SequencesSQL(schema string) string
	// UpdateSequenceSQL return SQL Command to update sequence to max +1  of tablename/column values
	UpdateSequenceSQL(schema string, sequence string, tableName string, column string) string
	StatusSequenceSQL(schema string, sequence string) string
}

// NewSQLUpdator creates a new SQL Updator.
func NewSQLUpdator(url string, schema string, dialect Dialect) *SQLUpdator {
	return &SQLUpdator{
		url:     url,
		schema:  schema,
		dialect: dialect,
	}
}

// Extract the sequences name from the data base
func (e *SQLUpdator) Extract() ([]string, *sequence.Error) {
	db, err := dburl.Open(e.url)
	if err != nil {
		return nil, &sequence.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, &sequence.Error{Description: err.Error()}
	}
	rows, err := db.Query(e.dialect.SequencesSQL(e.schema))
	if err != nil {
		return nil, &sequence.Error{Description: err.Error()}
	}

	var (
		sequenceName string
		result       []string
	)

	for rows.Next() {
		err := rows.Scan(&sequenceName)
		if err != nil {
			return nil, &sequence.Error{Description: err.Error()}
		}

		log.Debug().Str("sequence", sequenceName).Msg("find new sequence")
		result = append(result, sequenceName)
	}

	return result, nil
}

// Status get the current value of the sequence
func (e SQLUpdator) Status(seq sequence.Sequence) (sequence.Sequence, *sequence.Error) {
	db, err := dburl.Open(e.url)
	if err != nil {
		return seq, &sequence.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return seq, &sequence.Error{Description: err.Error()}
	}

	SQL := e.dialect.StatusSequenceSQL(e.schema, seq.Name)
	log.Debug().Str("sql", SQL).Msg("SQL to get status sequence")
	rows, err := db.Query(SQL)
	if err != nil {
		return seq, &sequence.Error{Description: err.Error()}
	}

	var status int

	for rows.Next() {
		err := rows.Scan(&status)
		if err != nil {
			return seq, &sequence.Error{Description: err.Error()}
		}

		seq.Value = status
	}
	return seq, nil
}

// Update sequence
func (e *SQLUpdator) Update(seqList []sequence.Sequence) *sequence.Error {
	db, err := dburl.Open(e.url)
	if err != nil {
		return &sequence.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return &sequence.Error{Description: err.Error()}
	}

	for _, seq := range seqList {
		SQL := e.dialect.UpdateSequenceSQL(e.schema, seq.Name, seq.Table, seq.Column)
		log.Debug().Str("sql", SQL).Msg("SQL to update sequence")
		_, err = db.Query(SQL)
		if err != nil {
			return &sequence.Error{Description: err.Error()}
		}
	}

	return nil
}
