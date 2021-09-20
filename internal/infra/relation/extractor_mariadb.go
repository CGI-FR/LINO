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

	// import mariadb connector
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/cgi-fr/lino/pkg/relation"
)

// NewMariadbExtractorFactory creates a new mariadb extractor factory.
func NewMariadbExtractorFactory() *MariadbExtractorFactory {
	return &MariadbExtractorFactory{}
}

// MariadbExtractorFactory exposes methods to create new Mariadb extractors.
type MariadbExtractorFactory struct{}

// New return a Mariadb extractor
func (e *MariadbExtractorFactory) New(url string, schema string) relation.Extractor {
	return NewSQLExtractor(url, schema, MariadbDialect{})
}

type MariadbDialect struct{}

func (d MariadbDialect) SQL(schema string) string {
	SQL := `
SELECT
    tc.constraint_name,
    tc.table_name,
    kcu.column_name,
    kcu.referenced_table_name AS foreign_table_name,
    kcu.referenced_column_name AS foreign_column_name
FROM
    information_schema.table_constraints AS tc
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
            `

	if schema != "" {
		SQL += fmt.Sprintf("AND tc.table_schema = '%s'", schema)
	}
	return SQL
}
