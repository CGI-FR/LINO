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

package pull

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cgi-fr/lino/internal/infra/commonsql"
	"github.com/cgi-fr/lino/pkg/pull"
	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
)

// SQLDataSource to read in the pull process.
type SQLDataSource struct {
	url     string
	schema  string
	dbx     *sqlx.DB
	db      *sql.DB
	dialect commonsql.Dialect
}

// Open a connection to the SQL DB
func (ds *SQLDataSource) Open() error {
	db, err := dburl.Open(ds.url)
	if err != nil {
		return err
	}

	ds.db = db

	u, err := dburl.Parse(ds.url)
	if err != nil {
		return err
	}

	ds.dbx = sqlx.NewDb(db, u.UnaliasedDriver)

	err = ds.dbx.Ping()
	if err != nil {
		return err
	}

	return nil
}

// OpenWithDB Open a connection with a given DB (for mock)
func (ds *SQLDataSource) OpenWithDB(db *sql.DB) error {
	ds.db = db

	u, err := dburl.Parse(ds.url)
	if err != nil {
		return err
	}

	ds.dbx = sqlx.NewDb(db, u.UnaliasedDriver)

	err = ds.dbx.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (ds *SQLDataSource) Read(source pull.Table, filter pull.Filter) (pull.RowSet, error) {
	reader, err := ds.RowReader(source, filter)
	if err != nil {
		return nil, err
	}

	result := pull.RowSet{}
	for reader.Next() {
		result = append(result, reader.Value())
	}

	if reader.Error() != nil {
		return result, fmt.Errorf("%w", reader.Error())
	}

	return result, nil
}

// RowReader iterate over rows in table with filter
// Version modifiée
// RowReader generates a SQL query for reading rows from a table with optional filtering and limiting.
func (ds *SQLDataSource) RowReader(source pull.Table, filter pull.Filter) (pull.RowReader, error) {
	// Get SELECT query and values
	values, sql := ds.GetSelectSQLAndValues(source, filter)

	// If log level is more than debug level, this function will log all SQL Query
	commonsql.LogSQLQuery(sql, values, ds.dialect)

	// Execute the SQL query and return the iterator
	rows, err := ds.dbx.Queryx(sql, values...)
	if err != nil {
		return nil, err
	}

	return &SQLDataIterator{rows, nil, nil}, nil
}

func (ds *SQLDataSource) GetSelectSQLAndValues(source pull.Table, filter pull.Filter) ([]interface{}, string) {
	sqlColumns := &strings.Builder{}

	// Build Columns clause *******************************************
	if pcols := source.Columns; len(pcols) > 0 && source.ExportMode != pull.ExportModeAll {
		for idx := int(0); idx < len(pcols); idx++ {
			if idx > 0 {
				sqlColumns.Write([]byte(", "))
			}
			sqlColumns.Write([]byte(" " + ds.dialect.Quote(pcols[idx].Name)))
		}
	} else {
		sqlColumns.Write([]byte("*"))
	}

	// Build WHERE clause ********************************************
	sqlWhere, values := commonsql.GetWhereSQLAndValues(filter.Values, filter.Where, ds.dialect)
	if len(sqlWhere) == 0 {
		sqlWhere = " 1=1 "
	}

	// If schema name is inclued in table name
	if strings.Contains(string(source.Name), ".") {
		parts := strings.Split(string(source.Name), ".")
		ds.schema = parts[0]
		source.Name = pull.TableName(parts[1])
	}

	// Assemble the builders in order using the existing method Select/SelectLimit
	var sql string
	if filter.Limit > 0 {
		sql = ds.dialect.SelectLimit(string(source.Name), ds.schema, sqlWhere, filter.Distinct, filter.Limit, sqlColumns.String())
	} else {
		sql = ds.dialect.Select(string(source.Name), ds.schema, sqlWhere, filter.Distinct, sqlColumns.String())
	}
	return values, sql
}

// Close a connection to the SQL DB
func (ds *SQLDataSource) Close() error {
	err := ds.dbx.Close()
	if err != nil {
		return err
	}
	return nil
}

// SQLDataIterator read data from a SQL database.
type SQLDataIterator struct {
	rows  *sqlx.Rows
	value pull.Row
	err   error
}

// Next reads the next rows if it exists.
func (di *SQLDataIterator) Next() bool {
	if di.rows == nil {
		return false
	}
	if di.rows.Next() {
		columns, err := di.rows.Columns()
		if err != nil {
			di.err = err
			return false
		}

		values, err := di.rows.SliceScan()
		if err != nil {
			di.err = err
			return false
		}

		row := pull.Row{}
		for i, column := range columns {
			row[column] = values[i]
		}
		di.value = row
		return true
	}
	if di.rows.Err() != nil {
		di.err = di.rows.Err()
	}
	return false
}

// Value returns the last read row.
func (di *SQLDataIterator) Value() pull.Row {
	return di.value
}

// Error returns the iterator error
func (di *SQLDataIterator) Error() error {
	return di.err
}

func NewSQLDataSource(url, schema string, dbx *sqlx.DB, db *sql.DB, dialect commonsql.Dialect) *SQLDataSource {
	return &SQLDataSource{
		url:     url,
		schema:  schema,
		dbx:     dbx,
		db:      db,
		dialect: dialect,
	}
}
