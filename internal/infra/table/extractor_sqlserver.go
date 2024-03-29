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

	_ "github.com/microsoft/go-mssqldb"

	"github.com/cgi-fr/lino/internal/infra/commonsql"
	"github.com/cgi-fr/lino/pkg/table"
)

// NewSQLServerExtractorFactory creates a new SQL Server extractor factory.
func NewSQLServerExtractorFactory() *SQLServerExtractorFactory {
	return &SQLServerExtractorFactory{}
}

// SQLServerExtractorFactory exposes methods to create new SQL Server extractors.
type SQLServerExtractorFactory struct{}

// New return a SQL Server extractor
func (e *SQLServerExtractorFactory) New(connectionString string, schema string) table.Extractor {
	return NewSQLExtractor(connectionString, schema, SQLServerDialect{commonsql.SQLServerDialect{}})
}

type SQLServerDialect struct {
	commonsql.Dialect
}

func (d SQLServerDialect) SQL(schema string) string {
	SQL := `SELECT kcu.table_schema,
	kcu.table_name,
	string_agg(kcu.column_name,',') AS key_columns
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

	SQL += `
GROUP BY tco.constraint_name,
	kcu.table_schema,
	kcu.table_name
ORDER BY kcu.table_schema,
	kcu.table_name`

	return SQL
}

func (d SQLServerDialect) GetExportType(dbtype string) (string, bool) {
	switch dbtype {
	// String types
	case "BPCHAR", "CHARACTER", "VARCHAR", "TEXT",
		"CHAR", "NCHAR", "NVARCHAR", "CLOB", "NCLOB", "NTEXT":
		return "string", true
	// Numeric types
	case "NUMERIC", "DECIMAL", "FLOAT", "REAL", "MONEY", "SMALLMONEY", "BIGINT",
		"INT", "TINYINT", "SMALLINT":
		return "numeric", true
	// Timestamp types
	case "TIMESTAMP", "TIMESTAMPTZ",
		"TIMESTAMP WITH TIME ZONE", "TIMESTAMP WITH LOCAL TIME ZONE":
		return "timestamp", true
	// Datetime types
	case "DATE", "DATETIME2", "SMALLDATETIME", "DATETIME", "TIME", "DATETIMEOFFSET":
		return "datetime", true
	// Binary types
	case "BYTEA",
		"BINARY", "VARBINARY", "IMAGE":
		return "base64", true
	default:
		return "", false
	}
}
