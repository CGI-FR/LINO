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
	"net/url"
	"slices"
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
	sqlLogger          *SQLLogger
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

func (dd *SQLDataDestination) SafeUrl() string {
	u, err := url.Parse(dd.url)
	if err != nil {
		return "invalid URL"
	}
	// remove usename and password
	u.User = nil
	return u.String()
}

// Close SQL connections
func (dd *SQLDataDestination) Close() *push.Error {
	errors := []*push.Error{}

	for _, rw := range dd.rowWriter {
		err := rw.close()
		if err != nil {
			log.Warn().Str("table", rw.table.Name()).AnErr("error", err).Msg("Error during row writer closing")
			errors = append(errors, err)
		}
	}

	err := dd.tx.Commit()
	if err != nil {
		errors = append(errors, &push.Error{Description: err.Error()})
	} else {
		log.Debug().Msg("transaction committed")
	}

	if dd.disableConstraints {
		for _, rw := range dd.rowWriter {
			err := rw.enableConstraints()
			if err != nil {
				log.Warn().Str("table", rw.table.Name()).AnErr("error", err).Msg("Error during constraints restore")
				errors = append(errors, err)
			}
		}
	}

	if err2 := dd.db.Close(); err2 != nil {
		log.Warn().AnErr("error", err2).Msg("Error during db closing")
		errors = append(errors, &push.Error{Description: err2.Error()})
	}

	if len(errors) > 0 {
		allErrors := &push.Error{}
		for _, err := range errors {
			allErrors.Description += "\n" + err.Description
		}
		return allErrors
	}

	return nil
}

// Commit SQL for connection
func (dd *SQLDataDestination) Commit() *push.Error {
	for _, rw := range dd.rowWriter {
		err := rw.close()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
	}

	err := dd.tx.Commit()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	log.Debug().Msg("transaction committed")

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

	dbx := sqlx.NewDb(db, u.UnaliasedDriver)

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

func (dd *SQLDataDestination) OpenSQLLogger(folderPath string) error {
	dd.sqlLogger = NewSQLLogger(folderPath)
	if err := dd.sqlLogger.Open(); err != nil {
		dd.sqlLogger = nil
		return &push.Error{Description: err.Error()}
	}

	return nil
}

// SQLRowWriter write data to a SQL table.
type SQLRowWriter struct {
	table               push.Table
	dd                  *SQLDataDestination
	duplicateKeysCache  map[push.Value]struct{}
	statement           *sql.Stmt
	headers             ValueHeaders
	disabledConstraints []SQLConstraint
	sqlLogger           *SQLLoggerWriter
}

// NewSQLRowWriter creates a new SQL row writer.
func NewSQLRowWriter(table push.Table, dd *SQLDataDestination) *SQLRowWriter {
	return &SQLRowWriter{
		table:               table,
		dd:                  dd,
		disabledConstraints: []SQLConstraint{},
	}
}

// open table writer
func (rw *SQLRowWriter) open() *push.Error {
	log.Debug().Msg(fmt.Sprintf("open table with mode %s", rw.dd.mode))
	if rw.dd.disableConstraints {
		err2 := rw.disableConstraints()
		if err2 != nil {
			return &push.Error{Description: err2.Error()}
		}
	}

	if rw.dd.mode == push.Truncate {
		err := rw.truncate()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
	}

	rw.duplicateKeysCache = map[push.Value]struct{}{}
	return nil
}

// close table writer
func (rw *SQLRowWriter) close() *push.Error {
	if rw.statement != nil {
		err := rw.statement.Close()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
		rw.statement = nil
		log.Debug().Msg(fmt.Sprintf("close statement %s", rw.dd.mode))
	}
	rw.sqlLogger.Close()
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

func (rw *SQLRowWriter) createStatement(row push.Row, where push.Row) *push.Error {
	if rw.statement != nil {
		return nil
	}

	selectValues, whereValues := rw.computeStatementInfos(row, where)

	var prepareStmt string
	var pusherr *push.Error

	log.Debug().Msg(fmt.Sprintf("received mode %s", rw.dd.mode))

	switch {
	case rw.dd.mode == push.Delete:
		/* #nosec */
		prepareStmt = "DELETE FROM " + rw.tableName() + " WHERE "
		for i := 0; i < len(whereValues); i++ {
			prepareStmt += whereValues[i].name + "=" + rw.dd.dialect.Placeholder(i+1)
			if i < len(whereValues)-1 {
				prepareStmt += " and "
			}
		}
		rw.headers = whereValues

	case rw.dd.mode == push.Update:
		prepareStmt, rw.headers, pusherr = rw.dd.dialect.UpdateStatement(rw.tableName(), selectValues, whereValues, rw.table.PrimaryKey())
		if pusherr != nil {
			return pusherr
		}

	case rw.dd.mode == push.Upsert:
		prepareStmt, rw.headers, pusherr = rw.dd.dialect.UpsertStatement(rw.tableName(), selectValues, whereValues, rw.table.PrimaryKey())
		if pusherr != nil {
			return pusherr
		}

	default: // Insert:
		prepareStmt, rw.headers = rw.dd.dialect.InsertStatement(rw.tableName(), selectValues, rw.table.PrimaryKey())
	}

	log.Debug().Stringer("headers", rw.headers).Msg(prepareStmt)

	stmt, err := rw.dd.tx.Prepare(prepareStmt)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	rw.statement = stmt
	rw.sqlLogger = rw.dd.sqlLogger.OpenWriter(rw.table, prepareStmt)
	return nil
}

type ValueDescriptor struct {
	name     string
	override bool // value in row is overridden (used for key translations)
	column   push.Column
}

type ValueHeaders []ValueDescriptor

func (vh ValueHeaders) String() string {
	sb := &strings.Builder{}
	sb.WriteString("[")
	for i, vd := range vh {
		sb.WriteString(vd.name)
		if i+1 < len(vh) {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

func (rw *SQLRowWriter) computeStatementInfos(row push.Row, where push.Row) (selectValues []ValueDescriptor, whereValues []ValueDescriptor) {
	for _, pk := range rw.table.PrimaryKey() {
		if _, ok := where[pk]; ok {
			whereValues = append(whereValues, ValueDescriptor{pk, true, rw.table.GetColumn(pk)})
		} else {
			whereValues = append(whereValues, ValueDescriptor{pk, false, rw.table.GetColumn(pk)})
		}
	}

	for k := range row {
		selectValues = append(selectValues, ValueDescriptor{k, false, rw.table.GetColumn(k)})
	}

	return
}

// Write
func (rw *SQLRowWriter) Write(row push.Row, where push.Row) *push.Error {
	err1 := rw.createStatement(row, where)
	if err1 != nil {
		return err1
	}

	importedRow, err15 := rw.table.Import(row)
	if err15 != nil {
		return err15
	}

	values := []interface{}{}
	for _, h := range rw.headers {
		if oldvalue, exists := where[h.name]; exists && h.override {
			values = append(values, rw.dd.dialect.ConvertValue(oldvalue, h))
		} else {
			values = append(values, rw.dd.dialect.ConvertValue(importedRow.GetOrNil(h.name), h))
		}
	}
	log.Trace().Stringer("headers", rw.headers).Str("table", rw.table.Name()).Msg(fmt.Sprint(values))

	rw.sqlLogger.Write(values)

	_, err2 := rw.statement.Exec(values...)
	log.Trace().AnErr("error", err2).Msg("push error")
	if err2 != nil {
		// reset statement after error
		if err := rw.close(); err != nil {
			return &push.Error{Description: err.Error() + "\noriginal error :\n" + err2.Error()}
		}
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

	if _, err := rw.dd.db.Exec(stm); err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

func (rw *SQLRowWriter) disableConstraints() *push.Error {
	if rw.dd.dialect.CanDisableIndividualConstraints() {
		readStm := rw.dd.dialect.ReadConstraintsStatement(rw.tableName())
		log.Debug().Msg(readStm)

		result, err := rw.dd.db.Query(readStm)
		if err != nil {
			return &push.Error{Description: err.Error()}
		}

		defer result.Close()

		var tableName, constraintName string
		for result.Next() {
			err := result.Scan(&tableName, &constraintName)
			if err != nil {
				return &push.Error{Description: err.Error()}
			}

			log.Info().Str("table", tableName).Str("constraint", constraintName).Msg("disabling constraint")
			stm := rw.dd.dialect.DisableConstraintStatement(tableName, constraintName)
			log.Debug().Msg(stm)

			if _, err := rw.dd.db.Exec(stm); err != nil {
				return &push.Error{Description: err.Error()}
			}

			rw.disabledConstraints = append(rw.disabledConstraints, SQLConstraint{tableName, constraintName})
		}
	} else {
		stm := rw.dd.dialect.DisableConstraintsStatement(rw.tableName())
		log.Debug().Msg(stm)

		if _, err := rw.dd.db.Exec(stm); err != nil {
			return &push.Error{Description: err.Error()}
		}
	}
	return nil
}

func (rw *SQLRowWriter) enableConstraints() *push.Error {
	if rw.dd.dialect.CanDisableIndividualConstraints() {
		for i := len(rw.disabledConstraints) - 1; i >= 0; i-- {
			constraint := rw.disabledConstraints[i]
			log.Info().Str("table", constraint.tableName).Str("constraint", constraint.constraintName).Msg("enabling constraint")

			stm := rw.dd.dialect.EnableConstraintStatement(constraint.tableName, constraint.constraintName)
			log.Debug().Msg(stm)

			if _, err := rw.dd.db.Exec(stm); err != nil {
				return &push.Error{Description: err.Error()}
			}
		}
		rw.disabledConstraints = []SQLConstraint{}
	} else {
		stm := rw.dd.dialect.EnableConstraintsStatement(rw.tableName())
		log.Debug().Msg(stm)

		if _, err := rw.dd.db.Exec(stm); err != nil {
			return &push.Error{Description: err.Error()}
		}
	}
	return nil
}

// isAPrimaryKey return true if columnName is in pknames
func isAPrimaryKey(columnName string, pkNames []string) bool {
	for _, pkName := range pkNames {
		if pkName == columnName {
			return true
		}
	}
	return false
}

// SQLDialect is an interface to inject SQL variations
type SQLDialect interface {
	Placeholder(int) string
	DisableConstraintsStatement(tableName string) string
	EnableConstraintsStatement(tableName string) string
	TruncateStatement(tableName string) string
	InsertStatement(tableName string, selectValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor)
	UpsertStatement(tableName string, selectValues []ValueDescriptor, whereValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor, err *push.Error)
	UpdateStatement(tableName string, selectValues []ValueDescriptor, whereValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor, err *push.Error)
	IsDuplicateError(error) bool
	ConvertValue(push.Value, ValueDescriptor) push.Value

	CanDisableIndividualConstraints() bool

	// ReadConstraintsStatement create a query that returns tableName and constraintName
	ReadConstraintsStatement(tableName string) string
	DisableConstraintStatement(tableName string, constraintName string) string
	EnableConstraintStatement(tableName string, constraintName string) string

	SupportPreserve() []string
	BlankTest(name string) string
	EmptyTest(name string) string
}

type SQLConstraint struct {
	tableName      string
	constraintName string
}

func appendColumnToSQL(column ValueDescriptor, sql *strings.Builder, d SQLDialect, index int) *push.Error {
	// check if preserve is in supported values

	preserveKind := push.PreserveNothing

	if column.column != nil {
		preserveKind = column.column.Preserve()
	}

	if !slices.Contains(d.SupportPreserve(), preserveKind) {
		return &push.Error{
			Description: fmt.Sprintf("Unsupported preserve value [%s] for column [%s]", column.column.Preserve(), column.name),
		}
	}

	switch {
	// preserve nothing
	case column.column == nil || column.column.Preserve() == push.PreserveNothing:
		sql.WriteString(column.name)
		sql.WriteString("=")
		sql.WriteString(d.Placeholder(index + 1))

	// preserve null
	case column.column.Preserve() == push.PreserveNull:
		sql.WriteString(column.name)
		sql.WriteString(" = CASE WHEN ")
		sql.WriteString(column.name)
		sql.WriteString(" IS NOT NULL THEN ")
		sql.WriteString(d.Placeholder(index + 1))
		sql.WriteString(" ELSE ")
		sql.WriteString(column.name)
		sql.WriteString(" END")
		// preserve empty string ""
	case column.column.Preserve() == push.PreserveEmpty:
		sql.WriteString(column.name)
		sql.WriteString(" = CASE WHEN ")
		sql.WriteString(d.EmptyTest(column.name))
		sql.WriteString(" THEN ")
		sql.WriteString(column.name)
		sql.WriteString(" ELSE ")
		sql.WriteString(d.Placeholder(index + 1))
		sql.WriteString(" END")
		// preserve empty string "" or null or all space string
	case column.column.Preserve() == push.PreserveBlank:
		sql.WriteString(column.name)
		sql.WriteString(" = CASE")
		sql.WriteString(" WHEN ")
		sql.WriteString(column.name)
		sql.WriteString(" IS NULL THEN ")
		sql.WriteString(column.name)
		sql.WriteString(" WHEN ")
		sql.WriteString(d.BlankTest(column.name))
		sql.WriteString(" THEN ")
		sql.WriteString(column.name)
		sql.WriteString(" ELSE ")
		sql.WriteString(d.Placeholder(index + 1))
		sql.WriteString(" END")

	default:
		return &push.Error{
			Description: fmt.Sprintf("Unsupported preserve value [%s] for column [%s]", column.column.Preserve(), column.name),
		}
	}

	return nil
}
