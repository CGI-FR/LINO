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
	"fmt"

	// import postgresql connector
	"github.com/cgi-fr/lino/pkg/sequence"
	_ "github.com/lib/pq"
)

// NewPostgresUpdatorFactory creates a new postgres extractor factory.
func NewPostgresUpdatorFactory() *PostgresUpdatorFactory {
	return &PostgresUpdatorFactory{}
}

// PostgresUpdatorFactory exposes methods to create new Postgres extractors.
type PostgresUpdatorFactory struct{}

// New return a Postgres extractor
func (e *PostgresUpdatorFactory) New(url string, schema string) sequence.Updator {
	return NewSQLUpdator(url, schema, PostgresDialect{})
}

type PostgresDialect struct{}

func (d PostgresDialect) SequencesSQL(schema string) string {
	SQL := "SELECT c.relname FROM pg_class c WHERE c.relkind = 'S'"

	if schema != "" {
		SQL += fmt.Sprintf(" AND  relnamespace::regnamespace::text  = '%s'", schema)
	}

	return SQL
}

func (d PostgresDialect) UpdateSequenceSQL(schema string, sequence string, tableName string, column string) string {
	return fmt.Sprintf("select setval('%s',  (SELECT GREATEST(MAX(%s), 1)  FROM %s));", sequence, column, tableName)
}

func (d PostgresDialect) StatusSequenceSQL(schema string, sequence string) string {
	return fmt.Sprintf("SELECT last_value FROM %s;", sequence)
}
