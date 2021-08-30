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

	"github.com/cgi-fr/lino/pkg/pull"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xo/dburl"
)

// SQLDataSource to read in the pull process.
type SQLDataSource struct {
	url     string
	schema  string
	dbx     *sqlx.DB
	db      *sql.DB
	dialect SQLDialect
}

// Open a connection to the SQL DB
func (ds *SQLDataSource) Open() *pull.Error {
	db, err := dburl.Open(ds.url)
	if err != nil {
		return &pull.Error{Description: err.Error()}
	}

	ds.db = db

	u, err := dburl.Parse(ds.url)
	if err != nil {
		return &pull.Error{Description: err.Error()}
	}

	ds.dbx = sqlx.NewDb(db, u.Unaliased)

	err = ds.dbx.Ping()
	if err != nil {
		return &pull.Error{Description: err.Error()}
	}

	return nil
}

// build table name with or without schema from dataconnector
func (ds *SQLDataSource) tableName(source pull.Table) string {
	if ds.schema == "" {
		return source.Name()
	}
	if strings.Contains(source.Name(), ".") {
		return source.Name()
	}
	return ds.schema + "." + source.Name()
}

// RowReader iterate over rows in table with filter
func (ds *SQLDataSource) RowReader(source pull.Table, filter pull.Filter) (pull.RowReader, *pull.Error) {
	sql := &strings.Builder{}
	sql.Write([]byte("SELECT * FROM "))
	sql.Write([]byte(ds.tableName(source)))
	sql.Write([]byte(" WHERE "))

	whereContentFlag := false

	values := []interface{}{}
	for key, value := range filter.Values() {
		sql.Write([]byte(key))
		values = append(values, value)
		fmt.Fprint(sql, "=")
		fmt.Fprint(sql, ds.dialect.Placeholder(len(values)))
		if len(values) < len(filter.Values()) {
			sql.Write([]byte(" AND "))
		}
		whereContentFlag = true
	}

	if filter.Where() != "" {
		if whereContentFlag {
			sql.Write([]byte(" AND "))
		}
		fmt.Fprint(sql, filter.Where())
		whereContentFlag = true
	}

	if !whereContentFlag {
		sql.Write([]byte(" 1=1 "))
	}

	if filter.Limit() > 0 {
		fmt.Fprint(sql, ds.dialect.Limit(filter.Limit()))
	}

	if log.Logger.GetLevel() <= zerolog.DebugLevel {
		printSQL := sql.String()
		for i, v := range values {
			printSQL = strings.ReplaceAll(printSQL, ds.dialect.Placeholder(i+1), fmt.Sprintf("%v", v))
		}
		log.Debug().Msg(fmt.Sprint(printSQL))
	}

	rows, err := ds.dbx.Queryx(sql.String(), values...)
	if err != nil {
		return nil, &pull.Error{Description: err.Error()}
	}

	return &SQLDataIterator{rows, nil, nil}, nil
}

// Close a connection to the SQL DB
func (ds *SQLDataSource) Close() *pull.Error {
	err := ds.dbx.Close()
	if err != nil {
		return &pull.Error{Description: err.Error()}
	}
	return nil
}

// SQLDataIterator read data from a SQL database.
type SQLDataIterator struct {
	rows  *sqlx.Rows
	value pull.Row
	err   *pull.Error
}

// Next reads the next rows if it exists.
func (di *SQLDataIterator) Next() bool {
	if di.rows == nil {
		return false
	}
	if di.rows.Next() {
		columns, err := di.rows.Columns()
		if err != nil {
			di.err = &pull.Error{Description: err.Error()}
			return false
		}

		values, err := di.rows.SliceScan()
		if err != nil {
			di.err = &pull.Error{Description: err.Error()}
			return false
		}

		row := pull.Row{}
		for i, column := range columns {
			b, ok := values[i].([]byte)
			if ok {
				row[column] = string(b)
			} else {
				row[column] = values[i]
			}
		}
		di.value = row
		return true
	}
	if di.rows.Err() != nil {
		di.err = &pull.Error{Description: di.rows.Err().Error()}
	}
	return false
}

// Value returns the last read row.
func (di *SQLDataIterator) Value() pull.Row {
	return di.value
}

// Error returns the iterator error
func (di *SQLDataIterator) Error() *pull.Error {
	return di.err
}

// SQLDialect to inject SQL variations
type SQLDialect interface {

	// Placeholder format variable in query
	Placeholder(int) string
	// Limit format limitation clause
	Limit(uint) string
}
