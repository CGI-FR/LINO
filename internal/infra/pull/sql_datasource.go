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
	"time"

	"github.com/cgi-fr/lino/internal/infra/commonsql"
	"github.com/cgi-fr/lino/pkg/pull"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xo/dburl"
)

func WithMaxLifetime(maxLifeTime time.Duration) pull.DataSourceOption {
	return func(ds pull.DataSource) {
		log.Info().Int64("maxLifetime", int64(maxLifeTime.Seconds())).Msg("setting database connection parameter")
		ds.(*SQLDataSource).maxLifetime = maxLifeTime
	}
}

func WithMaxOpenConns(maxOpenConns int) pull.DataSourceOption {
	return func(ds pull.DataSource) {
		log.Info().Int("maxOpenConns", maxOpenConns).Msg("setting database connection parameter")
		ds.(*SQLDataSource).maxOpenConns = maxOpenConns
	}
}

func WithMaxIdleConns(maxIdleConns int) pull.DataSourceOption {
	return func(ds pull.DataSource) {
		log.Info().Int("maxIdleConns", maxIdleConns).Msg("setting database connection parameter")
		ds.(*SQLDataSource).maxIdleConns = maxIdleConns
	}
}

// SQLDataSource to read in the pull process.
type SQLDataSource struct {
	url          string
	schema       string
	dbx          *sqlx.DB
	db           *sql.DB
	dialect      commonsql.Dialect
	maxLifetime  time.Duration
	maxOpenConns int
	maxIdleConns int
}

// Open a connection to the SQL DB
func (ds *SQLDataSource) Open() error {
	db, err := dburl.Open(ds.url)
	if err != nil {
		return err
	}

	log.Error().Msg("open database connection pool")

	// database handle settings
	db.SetConnMaxLifetime(ds.maxLifetime)
	db.SetMaxOpenConns(ds.maxOpenConns)
	db.SetMaxIdleConns(ds.maxIdleConns)

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

// build table name with or without schema from dataconnector
func (ds *SQLDataSource) tableName(source pull.Table) string {
	if ds.schema == "" {
		return string(source.Name)
	}
	if strings.Contains(string(source.Name), ".") {
		return string(source.Name)
	}
	return ds.schema + "." + string(source.Name)
}

func (ds *SQLDataSource) Read(source pull.Table, filter pull.Filter) (pull.RowSet, error) {
	reader, err := ds.RowReader(source, filter)
	if err != nil {
		return nil, err
	}

	defer reader.Close()

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
	values, sql := ds.GetSelectSQLAndValues(source, filter)

	if log.Logger.GetLevel() <= zerolog.DebugLevel {
		printSQL := sql
		for i, v := range values {
			printSQL = strings.ReplaceAll(printSQL, ds.dialect.Placeholder(i+1), fmt.Sprintf("%v", v))
		}
		log.Debug().Msg(fmt.Sprint(printSQL))
	}
	// Execute the SQL query and return the iterator
	rows, err := ds.dbx.Queryx(sql, values...)
	if err != nil {
		return nil, err
	}

	log.Error().Msg("open database rows iterator")

	return &SQLDataIterator{rows, nil, nil}, nil
}

func (ds *SQLDataSource) GetSelectSQLAndValues(source pull.Table, filter pull.Filter) ([]interface{}, string) {
	sqlWhere := &strings.Builder{}
	sqlColumns := &strings.Builder{}

	// Build Columns clause *******************************************
	if pcols := source.Columns; len(pcols) > 0 && source.ExportMode != pull.ExportModeAll {
		for idx := int(0); idx < len(pcols); idx++ {
			if idx > 0 {
				sqlColumns.Write([]byte(", "))
			}
			sqlColumns.Write([]byte(" " + pcols[idx].Name))
		}
	} else {
		sqlColumns.Write([]byte("*"))
	}

	// Build WHERE clause ********************************************
	whereContentFlag := false
	values := []interface{}{}
	for key, value := range filter.Values {
		sqlWhere.Write([]byte(key))
		values = append(values, value)
		fmt.Fprint(sqlWhere, "=")
		fmt.Fprint(sqlWhere, ds.dialect.Placeholder(len(values)))
		if len(values) < len(filter.Values) {
			sqlWhere.Write([]byte(" AND "))
		}
		whereContentFlag = true
	}

	if strings.TrimSpace(filter.Where) != "" {
		if whereContentFlag {
			sqlWhere.Write([]byte(" AND "))
		}
		fmt.Fprint(sqlWhere, filter.Where)
		whereContentFlag = true
	}

	if !whereContentFlag {
		sqlWhere.Write([]byte(" 1=1 "))
	}

	// Assemble the builders in order using the existing method Select/SelectLimit
	var sql string
	if filter.Limit > 0 {
		sql = ds.dialect.SelectLimit(ds.tableName(source), "", sqlWhere.String(), filter.Distinct, filter.Limit, sqlColumns.String())
	} else {
		sql = ds.dialect.Select(ds.tableName(source), "", sqlWhere.String(), filter.Distinct, sqlColumns.String())
	}
	return values, sql
}

// Close a connection to the SQL DB
func (ds *SQLDataSource) Close() error {
	err := ds.dbx.Close()
	if err != nil {
		return err
	}
	log.Error().Msg("close database connection pool")
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

// Close returns the iterator
func (di *SQLDataIterator) Close() error {
	defer log.Error().Msg("close database rows iterator")
	return di.rows.Close()
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
