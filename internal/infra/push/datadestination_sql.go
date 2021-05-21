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

package push

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cgi-fr/lino/pkg/push"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/xo/dburl"
)

// SQLDataDestination read data from a SQL database.
type SQLDataDestination struct {
	url                string
	schema             string
	db                 *sqlx.DB
	tx                 *sql.Tx
	rowWriter          map[string]*SQLRowWriter
	mode               push.Mode
	disableConstraints bool
	dialect            SQLDialect
}

// NewSQLDataDestination creates a new SQL datadestination.
func NewSQLDataDestination(url string, schema string, dialect SQLDialect) *SQLDataDestination {
	return &SQLDataDestination{
		url:       url,
		schema:    schema,
		rowWriter: map[string]*SQLRowWriter{},
		dialect:   dialect,
	}
}

// Close SQL connections
func (dd *SQLDataDestination) Close() *push.Error {
	for _, rw := range dd.rowWriter {
		err := rw.commit()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
	}

	err := dd.tx.Commit()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	log.Debug().Msg("transaction committed")

	for _, rw := range dd.rowWriter {
		err := rw.close()
		if err != nil {
			return err
		}
	}

	err2 := dd.db.Close()
	if err2 != nil {
		return &push.Error{Description: err2.Error()}
	}

	return nil
}

// Commit SQL for connection
func (dd *SQLDataDestination) Commit() *push.Error {
	for _, rw := range dd.rowWriter {
		err := rw.commit()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
	}

	err := dd.tx.Commit()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	log.Debug().Msg("transaction committed")

	for _, rw := range dd.rowWriter {
		err := rw.commit()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
	}

	tx, err := dd.db.Begin()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}

	dd.tx = tx

	return nil
}

// Open SQL Connection
func (dd *SQLDataDestination) Open(plan push.Plan, mode push.Mode, disableConstraints bool) *push.Error {
	dd.mode = mode
	dd.disableConstraints = disableConstraints

	db, err := dburl.Open(dd.url)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}

	u, err := dburl.Parse(dd.url)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}

	dbx := sqlx.NewDb(db, u.Unaliased)

	err = dbx.Ping()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}

	dd.db = dbx

	tx, err := dd.db.Begin()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	dd.tx = tx

	for _, table := range plan.Tables() {
		rw := NewSQLRowWriter(table, dd)
		err := rw.open()
		if err != nil {
			return err
		}

		dd.rowWriter[table.Name()] = rw
	}

	return nil
}

// RowWriter return SQL table writer
func (dd *SQLDataDestination) RowWriter(table push.Table) (push.RowWriter, *push.Error) {
	rw, ok := dd.rowWriter[table.Name()]
	if ok {
		return rw, nil
	}

	rw = NewSQLRowWriter(table, dd)
	err := rw.open()
	if err != nil {
		return nil, err
	}

	dd.rowWriter[table.Name()] = rw
	return rw, nil
}

// SQLRowWriter write data to a SQL table.
type SQLRowWriter struct {
	table              push.Table
	dd                 *SQLDataDestination
	duplicateKeysCache map[push.Value]struct{}
	statement          *sql.Stmt
	headers            []string
}

// NewSQLRowWriter creates a new SQL row writer.
func NewSQLRowWriter(table push.Table, dd *SQLDataDestination) *SQLRowWriter {
	return &SQLRowWriter{
		table: table,
		dd:    dd,
	}
}

// open table writer
func (rw *SQLRowWriter) open() *push.Error {
	log.Debug().Msg(fmt.Sprintf("open table with mode %s", rw.dd.mode))
	if rw.dd.mode == push.Truncate {
		err := rw.truncate()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
	}

	if rw.dd.disableConstraints {
		err2 := rw.disableConstraints()
		if err2 != nil {
			return &push.Error{Description: err2.Error()}
		}
	}
	rw.duplicateKeysCache = map[push.Value]struct{}{}

	return nil
}

func (rw *SQLRowWriter) commit() *push.Error {
	if rw.statement != nil {
		err := rw.statement.Close()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
		rw.statement = nil
		log.Debug().Msg(fmt.Sprintf("close statement %s", rw.dd.mode))
	}
	return nil
}

// close table writer
func (rw *SQLRowWriter) close() *push.Error {
	if rw.statement != nil {
		err := rw.statement.Close()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
	}

	if rw.dd.disableConstraints {
		return rw.enableConstraints()
	}
	return nil
}

// build table name with or without schema from dataconnector
func (rw *SQLRowWriter) tableName() string {
	if rw.dd.schema == "" {
		return rw.table.Name()
	}
	if strings.Contains(rw.table.Name(), ".") {
		return rw.table.Name()
	}
	return rw.dd.schema + "." + rw.table.Name()
}

func (rw *SQLRowWriter) createStatement(row push.Row) *push.Error {
	if rw.statement != nil {
		return nil
	}

	names := []string{}
	valuesVar := []string{}
	pkNames := []string{}
	pkVar := []string{}

	i := 1
	for k := range row {
		names = append(names, k)
		valuesVar = append(valuesVar, rw.dd.dialect.Placeholder(i))
		for _, pk := range rw.table.PrimaryKey() {
			if pk == k {
				pkNames = append(pkNames, k)
				pkVar = append(pkVar, rw.dd.dialect.Placeholder(i))
			}
		}
		i++
	}

	var prepareStmt string
	var pusherr *push.Error
	log.Debug().Msg(fmt.Sprintf("received mode %s", rw.dd.mode))
	switch {
	case rw.dd.mode == push.Delete:
		/* #nosec */
		prepareStmt = "DELETE FROM " + rw.tableName() + " WHERE "
		for i := 0; i < len(pkNames); i++ {
			prepareStmt += pkNames[i] + "=" + rw.dd.dialect.Placeholder(i+1)
			if i < len(pkNames)-1 {
				prepareStmt += " and "
			}
		}
		rw.headers = pkNames
	case rw.dd.mode == push.Update:
		prepareStmt, pusherr = rw.dd.dialect.UpdateStatement(rw.tableName(), names, valuesVar, rw.table.PrimaryKey(), pkVar)
		if pusherr != nil {
			return pusherr
		}
		rw.headers = names
	default: // Insert:
		/* #nosec */
		prepareStmt = rw.dd.dialect.InsertStatement(rw.tableName(), names, valuesVar, rw.table.PrimaryKey())
		rw.headers = names
	}
	log.Debug().Msg(prepareStmt)

	stmt, err := rw.dd.tx.Prepare(prepareStmt)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	rw.statement = stmt
	return nil
}

// Write
func (rw *SQLRowWriter) Write(row push.Row) *push.Error {
	err1 := rw.createStatement(row)
	if err1 != nil {
		return err1
	}

	values := []interface{}{}
	for _, h := range rw.headers {
		values = append(values, rw.dd.dialect.ConvertValue(row[h]))
	}
	log.Trace().Msg(fmt.Sprint(values))

	_, err2 := rw.statement.Exec(values...)
	if err2 != nil {
		if rw.dd.dialect.IsDuplicateError(err2) {
			log.Trace().Msg(fmt.Sprintf("duplicate key %v (%s) for %s", row, rw.table.PrimaryKey(), rw.table.Name()))
		} else {
			return &push.Error{Description: err2.Error()}
		}
	}

	return nil
}

func (rw *SQLRowWriter) truncate() *push.Error {
	stm := rw.dd.dialect.TruncateStatement(rw.tableName())
	log.Debug().Msg(stm)
	_, err := rw.dd.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

func (rw *SQLRowWriter) disableConstraints() *push.Error {
	stm := rw.dd.dialect.DisableConstraintsStatement(rw.tableName())
	log.Debug().Msg(stm)
	_, err := rw.dd.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

func (rw *SQLRowWriter) enableConstraints() *push.Error {
	stm := rw.dd.dialect.EnableConstraintsStatement(rw.tableName())
	log.Debug().Msg(stm)
	_, err := rw.dd.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

// SQLDialect is an interface to inject SQL variations
type SQLDialect interface {
	Placeholder(int) string
	DisableConstraintsStatement(tableName string) string
	EnableConstraintsStatement(tableName string) string
	TruncateStatement(tableName string) string
	InsertStatement(tableName string, columns []string, values []string, primaryKeys []string) string
	UpdateStatement(tableName string, columns []string, uValues []string, primaryKeys []string, pValues []string) (string, *push.Error)
	IsDuplicateError(error) bool
	ConvertValue(push.Value) push.Value
}
