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

//go:build db2
// +build db2

package table

import (
	"fmt"

	// import db2 connector
	_ "github.com/ibmdb/go_ibm_db"

	"github.com/cgi-fr/lino/internal/infra/commonsql"
	"github.com/cgi-fr/lino/pkg/table"
)

// NewDb2ExtractorFactory creates a new postgres extractor factory.
func NewDb2ExtractorFactory() *Db2ExtractorFactory {
	return &Db2ExtractorFactory{}
}

// Db2ExtractorFactory exposes methods to create new Db2 extractors.
type Db2ExtractorFactory struct{}

// New return a Db2 extractor
func (e *Db2ExtractorFactory) New(url string, schema string) table.Extractor {
	return NewSQLExtractor(url, schema, Db2Dialect{commonsql.Db2Dialect{}})
}

type Db2Dialect struct {
	commonsql.Dialect
}

func (d Db2Dialect) SQL(schema string) string {
	SQL := `
	select
		tab.tabschema as schema_name,
		tab.tabname as table_name,
		LISTAGG(key.colname, ',') AS column_names
	from syscat.tables tab
	inner join syscat.tabconst const
		on const.tabschema = tab.tabschema
		and const.tabname = tab.tabname and const.type = 'P'
	inner join syscat.keycoluse key
		on const.tabschema = key.tabschema
		and const.tabname = key.tabname
		and const.constname = key.constname
	where tab.type = 'T'
		and tab.tabschema not like 'SYS%'
	group by tab.tabschema, tab.tabname
	`

	if schema != "" {
		SQL += fmt.Sprintf("and t.tabschema = '%s'", schema)
	}

	return SQL
}

func (d Db2Dialect) GetExportType(dbtype string) (string, bool) {
	switch dbtype {
	// String types
	case "TSVECTOR", "_TEXT", "BPCHAR", "CHARACTER", "CHARACTER VARYING", "VARCHAR", "TEXT",
		"CHAR", "VARCHAR2", "NCHAR", "NVARCHAR2", "CLOB", "NCLOB",
		"TINYTEXT", "MEDIUMTEXT", "LONGTEXT":
		return "string", true
	// Numeric types
	case "NUMERIC", "DECIMAL", "FLOAT", "REAL", "DOUBLE PRECISION", "MONEY", "INTEGER", "BIGINT",
		"NUMBER", "BINARY_FLOAT", "BINARY_DOUBLE", "INT", "TINYINT", "SMALLINT", "MEDIUMINT":
		return "numeric", true
	// Timestamp types
	case "TIMESTAMP", "TIMESTAMPTZ",
		"TIMESTAMP WITH TIME ZONE", "TIMESTAMP WITH LOCAL TIME ZONE":
		return "timestamp", true
	// Datetime types
	case "DATE", "DATETIME2", "SMALLDATETIME", "DATETIME":
		return "datetime", true
	// Binary types
	case "BYTEA",
		"RAW", "LONG RAW", "BINARY", "VARBINARY", "TINYBLOB", "MEDIUMBLOB", "LONGBLOB", "IMAGE", "BLOB":
		return "base64", true
	default:
		return "", false
	}
}
