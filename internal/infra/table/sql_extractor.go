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
	"database/sql"
	"strings"

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
	SQL := e.dialect.SQL(e.schema)

	rows, err := db.Query(SQL)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	tables := []table.Table{}

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
		// Get columns information, check is there have types needs to be modify in export
		columns, err := ColumnInfo(db, tableName)
		if err != nil {
			return nil, &table.Error{Description: err.Error()}
		}

		if len(columns) > 0 {
			table := table.Table{
				Name:       tableName,
				Keys:       strings.Split(keyColumns, ","),
				Columns:    columns,
				ExportMode: table.ExportModeAll,
			}

			tables = append(tables, table)
		} else {
			table := table.Table{
				Name: tableName,
				Keys: strings.Split(keyColumns, ","),
			}

			tables = append(tables, table)
		}

	}
	err = rows.Err()
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	return tables, nil
}

func (e *SQLExtractor) Count(tableName string) (int, *table.Error) {
	db, err := dburl.Open(e.url)
	if err != nil {
		return 0, &table.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return 0, &table.Error{Description: err.Error()}
	}

	SQL := `SELECT COUNT(*) FROM ` + tableName

	rows, err := db.Query(SQL)
	if err != nil {
		return 0, &table.Error{Description: err.Error()}
	}

	var count int

	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, &table.Error{Description: err.Error()}
		}
	}

	err = rows.Err()
	if err != nil {
		return 0, &table.Error{Description: err.Error()}
	}

	return count, nil
}

func ColumnInfo(db *sql.DB, tableName string) ([]table.Column, error) {
	// Execute query to fetch column information
	rows, err := db.Query("SELECT * FROM " + tableName + " LIMIT 1")
	if err != nil {
		return []table.Column{}, err
	}
	defer rows.Close()

	// Retrieve column information
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		rows.Close()
		return []table.Column{}, err
	}

	columns := []table.Column{}
	// Iterate over column information
	for _, ct := range columnTypes {
		columnName := ct.Name()
		dataType := ct.DatabaseTypeName()

		exportType, save := checkType(dataType)
		if save {
			// columnLength, _ := ct.Length()
			// columnPrecision, columnSize, _ := ct.DecimalSize()

			columnInfo := table.Column{
				Name:   columnName,
				Export: exportType,
			}
			// if columnLength > 0 {
			// 	fmt.Printf(", Length: %d", columnLength)
			// } else if columnSize > 0 {
			// 	fmt.Printf(", Size: %d", columnSize)
			// 	fmt.Printf(", Precision: %d", columnPrecision)
			// }
			columns = append(columns, columnInfo)
		}
	}

	return columns, nil
}

func checkType(columnType string) (string, bool) {
	switch columnType {
	// string case
	case "TSVECTOR", "_TEXT", "", "BPCHAR":
		return "string", true
	// numeric case
	case "NUMERIC":
		return "numeric", true
	// timestamp case
	case "TIMESTAMP":
		return "timestamp", true
	// datetime case
	case "DATE":
		return "datetime", true
	// base64/binary case
	case "BYTEA":
		return "base64", true
	default:
		return "", false
	}
}
