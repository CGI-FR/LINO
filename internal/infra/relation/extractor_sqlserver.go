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
	"fmt"

	// import SQL Server connector
	_ "github.com/microsoft/go-mssqldb"

	"github.com/cgi-fr/lino/pkg/relation"
)

// NewSQLServerExtractorFactory creates a new SQL Server extractor factory.
func NewSQLServerExtractorFactory() *SQLServerExtractorFactory {
	return &SQLServerExtractorFactory{}
}

// SQLServerExtractorFactory exposes methods to create new SQL Server extractors.
type SQLServerExtractorFactory struct{}

// New return a SQL Server extractor
func (e *SQLServerExtractorFactory) New(connectionString string, schema string) relation.Extractor {
	return NewSQLExtractor(connectionString, schema, SQLServerDialect{})
}

type SQLServerDialect struct{}

func (d SQLServerDialect) SQL(schema string) string {
	SQL := `
SELECT
    tc.constraint_name,
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM
    information_schema.table_constraints AS tc
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
      AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
`

	if schema != "" {
		SQL += fmt.Sprintf("AND tc.table_schema = '%s'", schema)
	}
	return SQL
}
