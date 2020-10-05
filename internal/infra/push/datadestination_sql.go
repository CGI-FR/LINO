package push

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/push"
)

// SQLDataDestination read data from a SQL database.
type SQLDataDestination struct {
	url                string
	schema             string
	logger             push.Logger
	db                 *sqlx.DB
	rowWriter          map[string]*SQLRowWriter
	mode               push.Mode
	disableConstraints bool
	dialect            SQLDialect
}

// NewSQLDataDestination creates a new SQL datadestination.
func NewSQLDataDestination(url string, schema string, dialect SQLDialect, logger push.Logger) *SQLDataDestination {
	return &SQLDataDestination{
		url:       url,
		schema:    schema,
		logger:    logger,
		rowWriter: map[string]*SQLRowWriter{},
		dialect:   dialect,
	}
}

// Close SQL connections
func (dd *SQLDataDestination) Close() *push.Error {
	for _, rw := range dd.rowWriter {
		rw.close()
	}

	err := dd.db.Close()
	if err != nil {
		return &push.Error{Description: err.Error()}
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

	return nil
}

// Open SQL Connection
func (dd *SQLDataDestination) Open(plan push.Plan, mode push.Mode, disableConstraints bool) *push.Error {
	dd.mode = mode
	dd.disableConstraints = disableConstraints

	dd.logger.Info(fmt.Sprintf("connecting to %s...", dd.url))
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
	tx                 *sql.Tx
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
	rw.dd.logger.Debug(fmt.Sprintf("open table with mode %s", rw.dd.mode))
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

	err3 := rw.begin()
	if err3 != nil {
		return &push.Error{Description: err3.Error()}
	}
	return nil
}

func (rw *SQLRowWriter) begin() *push.Error {
	tx, err := rw.dd.db.Begin()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	rw.tx = tx
	return nil
}

func (rw *SQLRowWriter) commit() *push.Error {
	if rw.statement != nil {
		err := rw.statement.Close()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
		rw.statement = nil
	}
	if rw.tx != nil {
		err := rw.tx.Commit()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
		rw.tx = nil
	}
	return rw.begin()
}

// close table writer
func (rw *SQLRowWriter) close() *push.Error {
	if rw.statement != nil {
		err := rw.statement.Close()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
	}
	if rw.tx != nil {
		err := rw.tx.Commit()
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
	pkVar := []string{}

	i := 1
	for k := range row {
		names = append(names, k)
		valuesVar = append(valuesVar, rw.dd.dialect.Placeholder(i))
		for _, pk := range rw.table.PrimaryKey() {
			if pk == k {
				pkVar = append(pkVar, rw.dd.dialect.Placeholder(i))
			}
		}
		i++
	}

	var prepareStmt string
	var pusherr *push.Error
	rw.dd.logger.Debug(fmt.Sprintf("received mode %s", rw.dd.mode))
	switch {
	case rw.dd.mode == push.Delete:
		/* #nosec */
		prepareStmt = "DELETE FROM " + rw.tableName() + " WHERE "
		for i := 0; i < len(names); i++ {
			prepareStmt += names[i] + "=" + valuesVar[i]
			if i < len(names)-1 {
				prepareStmt += " and "
			}
		}
	case rw.dd.mode == push.Update:
		prepareStmt, pusherr = rw.dd.dialect.UpdateStatement(rw.tableName(), names, valuesVar, rw.table.PrimaryKey(), pkVar)
		if pusherr != nil {
			return pusherr
		}
	default: //Insert:
		/* #nosec */
		prepareStmt = rw.dd.dialect.InsertStatement(rw.tableName(), names, valuesVar, rw.table.PrimaryKey())
	}
	rw.dd.logger.Debug(prepareStmt)

	stmt, err := rw.tx.Prepare(prepareStmt)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	rw.statement = stmt
	rw.headers = names
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
	rw.dd.logger.Trace(fmt.Sprint(values))

	_, err2 := rw.statement.Exec(values...)
	if err2 != nil {
		if rw.dd.dialect.IsDuplicateError(err2) {
			rw.dd.logger.Trace(fmt.Sprintf("duplicate key %v (%s) for %s", row, rw.table.PrimaryKey(), rw.table.Name()))
		} else {
			return &push.Error{Description: err2.Error()}
		}
	}

	return nil
}

func (rw *SQLRowWriter) truncate() *push.Error {
	stm := rw.dd.dialect.TruncateStatement(rw.tableName())
	rw.dd.logger.Debug(stm)
	_, err := rw.dd.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

func (rw *SQLRowWriter) disableConstraints() *push.Error {
	stm := rw.dd.dialect.DisableConstraintsStatement(rw.tableName())
	rw.dd.logger.Debug(stm)
	_, err := rw.dd.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

func (rw *SQLRowWriter) enableConstraints() *push.Error {
	stm := rw.dd.dialect.EnableConstraintsStatement(rw.tableName())
	rw.dd.logger.Debug(stm)
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
