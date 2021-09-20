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
	"fmt"

	// import mariadbsql connector
	_ "github.com/go-sql-driver/mysql"

	"github.com/cgi-fr/lino/pkg/table"
)

// NewMariadbExtractorFactory creates a new mariadb extractor factory.
func NewMariadbExtractorFactory() *MariadbExtractorFactory {
	return &MariadbExtractorFactory{}
}

// MariadbExtractorFactory exposes methods to create new Mariadb extractors.
type MariadbExtractorFactory struct{}

// New return a Mariadb extractor
func (e *MariadbExtractorFactory) New(url string, schema string) table.Extractor {
	return NewSQLExtractor(url, schema, MariadbDialect{})
}

type MariadbDialect struct {
}

func (d MariadbDialect) SQL(schema string) string {
	SQL := `SELECT kcu.table_schema,
			kcu.table_name,
			GROUP_CONCAT(DISTINCT kcu.column_name SEPARATOR ',') AS key_columns
			FROM information_schema.table_constraints tco
			JOIN information_schema.key_column_usage kcu
			  ON kcu.constraint_name = tco.constraint_name
			  AND kcu.constraint_schema = tco.constraint_schema
              AND kcu.constraint_name = tco.constraint_name
			WHERE tco.constraint_type = 'PRIMARY KEY'
			`

	if schema != "" {
		SQL += fmt.Sprintf("AND kcu.table_schema = '%s'", schema)
	}

	SQL += `GROUP BY tco.constraint_name,
			kcu.table_schema,
			kcu.table_name
			ORDER BY kcu.table_schema,
			kcu.table_name`

	return SQL
}
