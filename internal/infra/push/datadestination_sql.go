package push

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/push"
)

// SQLDataDestination read data from a SQL database.
type SQLDataDestination struct {
	url       string
	logger    push.Logger
	db        *sqlx.DB
	rowWriter map[string]*SQLRowWriter
	mode      push.Mode
	dialect   SQLDialect
}

// NewSQLDataDestination creates a new SQL datadestination.
func NewSQLDataDestination(url string, dialect SQLDialect, logger push.Logger) *SQLDataDestination {
	return &SQLDataDestination{
		url:       url,
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

	err := dd.db.Close()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

// Open SQL Connection
func (dd *SQLDataDestination) Open(plan push.Plan, mode push.Mode) *push.Error {
	dd.mode = mode

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

	rw = NewSQLRowWriter(table, dd) //TODO
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

	err2 := rw.disableConstraints()
	if err2 != nil {
		return &push.Error{Description: err2.Error()}
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
	err := rw.tx.Commit()
	if err != nil {
		return &push.Error{Description: err.Error()}
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

	return rw.enableConstraints()
}

func (rw *SQLRowWriter) createStatement(row push.Row) *push.Error {
	if rw.statement != nil {
		return nil
	}

	names := []string{}
	valuesVar := []string{}

	i := 1
	for k := range row {
		names = append(names, k)
		valuesVar = append(valuesVar, rw.dd.dialect.Placeholder(i))
		i++
	}

	var prepareStmt string
	if rw.dd.mode == push.Delete {
		/* #nosec */
		prepareStmt = "DELETE FROM " + rw.table.Name() + " WHERE "
		for i := 0; i < len(names); i++ {
			prepareStmt += names[i] + "=" + valuesVar[i]
			if i < len(names)-1 {
				prepareStmt += " and "
			}
		}
	} else {
		/* #nosec */
		prepareStmt = "INSERT INTO " + rw.table.Name() + "(" + strings.Join(names, ",") + ") VALUES(" + strings.Join(valuesVar, ",") + ")"
	}
	rw.dd.logger.Debug(prepareStmt)
	// TODO: Create an update statement

	stmt, err := rw.dd.db.Prepare(prepareStmt)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	rw.statement = stmt
	rw.headers = names
	return nil
}

func (rw *SQLRowWriter) addToCache(row push.Row) {
	rw.duplicateKeysCache[rw.computeRowSum(row)] = struct{}{}
}

func (rw *SQLRowWriter) isInCache(row push.Row) bool {
	_, ok := rw.duplicateKeysCache[rw.computeRowSum(row)]
	return ok
}

func (rw *SQLRowWriter) computeRowSum(row push.Row) string {
	sum := ""
	for _, pk := range rw.table.PrimaryKey() {
		sum = fmt.Sprintf("%s|,-%v", sum, row[pk])
	}
	return sum
}

// Write
func (rw *SQLRowWriter) Write(row push.Row) *push.Error {
	if ok := rw.isInCache(row); ok {
		rw.dd.logger.Trace(fmt.Sprintf("duplicate key in dataset %v (%s) for %s", row, rw.table.PrimaryKey(), rw.table.Name()))
		return nil
	}
	rw.addToCache(row)

	err1 := rw.createStatement(row)
	if err1 != nil {
		return err1
	}

	values := []interface{}{}
	for _, h := range rw.headers {
		values = append(values, row[h])
	}
	rw.dd.logger.Trace(fmt.Sprint(values))

	_, err2 := rw.statement.Exec(values...)
	if err2 != nil {
		pqErr := err2.(*pq.Error)
		if pqErr.Code == "23505" { //duplicate
			rw.dd.logger.Trace(fmt.Sprintf("duplicate key %v (%s) for %s", row, rw.table.PrimaryKey(), rw.table.Name()))
			// TODO update
		} else {
			return &push.Error{Description: err2.Error()}
		}
	}

	return nil
}

func (rw *SQLRowWriter) truncate() *push.Error {
	stm := rw.dd.dialect.TruncateStatement(rw.table.Name())
	rw.dd.logger.Debug(stm)
	_, err := rw.dd.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

func (rw *SQLRowWriter) disableConstraints() *push.Error {
	stm := rw.dd.dialect.DisableConstraintsStatement(rw.table.Name())
	rw.dd.logger.Debug(stm)
	_, err := rw.dd.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

func (rw *SQLRowWriter) enableConstraints() *push.Error {
	stm := rw.dd.dialect.EnableConstraintsStatement(rw.table.Name())
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
}
