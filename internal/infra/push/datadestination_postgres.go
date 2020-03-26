package push

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/push"

	// import postgresql connector
	"github.com/lib/pq"
)

// PostgresDataDestinationFactory exposes methods to create new Postgres pullers.
type PostgresDataDestinationFactory struct {
	logger push.Logger
}

// NewPostgresDataDestinationFactory creates a new postgres datadestination factory.
func NewPostgresDataDestinationFactory(l push.Logger) *PostgresDataDestinationFactory {
	return &PostgresDataDestinationFactory{l}
}

// New return a Postgres pusher
func (e *PostgresDataDestinationFactory) New(url string) push.DataDestination {
	return NewPostgresDataDestination(url, e.logger)
}

// PostgresDataDestination read data from a PostgreSQL database.
type PostgresDataDestination struct {
	url       string
	logger    push.Logger
	db        *sqlx.DB
	rowWriter map[string]*PostgresRowWriter
	mode      push.Mode
}

// NewPostgresDataDestination creates a new postgres datadestination.
func NewPostgresDataDestination(url string, logger push.Logger) *PostgresDataDestination {
	return &PostgresDataDestination{
		url:       url,
		logger:    logger,
		rowWriter: map[string]*PostgresRowWriter{},
	}
}

// Close postgres connections
func (ds *PostgresDataDestination) Close() *push.Error {
	for _, rw := range ds.rowWriter {
		rw.close()
	}

	err := ds.db.Close()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

// Open postgres Connections
func (ds *PostgresDataDestination) Open(plan push.Plan, mode push.Mode) *push.Error {
	ds.mode = mode

	ds.logger.Info(fmt.Sprintf("connecting to %s...", ds.url))
	db, err := dburl.Open(ds.url)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}

	u, err := dburl.Parse(ds.url)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}

	dbx := sqlx.NewDb(db, u.Unaliased)

	err = dbx.Ping()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}

	ds.db = dbx

	for _, table := range plan.Tables() {
		rw := NewPostgresRowWriter(table, ds)
		err := rw.open()
		if err != nil {
			return err
		}

		ds.rowWriter[table.Name()] = rw
	}

	return nil
}

// RowWriter return postgres table writer
func (ds *PostgresDataDestination) RowWriter(table push.Table) (push.RowWriter, *push.Error) {
	rw, ok := ds.rowWriter[table.Name()]
	if ok {
		return rw, nil
	}

	rw = NewPostgresRowWriter(table, ds) //TODO
	err := rw.open()
	if err != nil {
		return nil, err
	}

	ds.rowWriter[table.Name()] = rw
	return rw, nil
}

// PostgresRowWriter write data to a PostgreSQL table.
type PostgresRowWriter struct {
	table              push.Table
	ds                 *PostgresDataDestination
	duplicateKeysCache map[push.Value]struct{}
	statement          *sql.Stmt
	headers            []string
}

// NewPostgresRowWriter creates a new postgres row writer.
func NewPostgresRowWriter(table push.Table, ds *PostgresDataDestination) *PostgresRowWriter {
	return &PostgresRowWriter{
		table: table,
		ds:    ds,
	}
}

// open table writer
func (rw *PostgresRowWriter) open() *push.Error {
	rw.ds.logger.Debug(fmt.Sprintf("open table with mode %s", rw.ds.mode))
	if rw.ds.mode == push.Truncate {
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
	return nil
}

// close table writer
func (rw *PostgresRowWriter) close() *push.Error {
	if rw.statement != nil {
		err := rw.statement.Close()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
	}

	return rw.enableConstraints()
}

func (rw *PostgresRowWriter) createStatement(row push.Row) *push.Error {
	if rw.statement != nil {
		return nil
	}

	names := []string{}
	valuesVar := []string{}

	i := 1
	for k := range row {
		names = append(names, k)
		valuesVar = append(valuesVar, fmt.Sprintf("$%d", i))
		i++
	}

	var prepareStmt string
	if rw.ds.mode == push.Delete {
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
	rw.ds.logger.Debug(prepareStmt)
	// TODO: Create an update statement

	stmt, err := rw.ds.db.Prepare(prepareStmt)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	rw.statement = stmt
	rw.headers = names
	return nil
}

// Write
func (rw *PostgresRowWriter) Write(row push.Row) *push.Error {
	if _, ok := rw.duplicateKeysCache[row[rw.table.PrimaryKey()]]; ok {
		rw.ds.logger.Trace(fmt.Sprintf("duplicate key in dataset %v (%s) for %s", row[rw.table.PrimaryKey()], rw.table.PrimaryKey(), rw.table.Name()))
		return nil
	}
	rw.duplicateKeysCache[row[rw.table.PrimaryKey()]] = struct{}{}

	err1 := rw.createStatement(row)
	if err1 != nil {
		return err1
	}

	values := []interface{}{}
	for _, h := range rw.headers {
		values = append(values, row[h])
	}
	rw.ds.logger.Trace(fmt.Sprint(values))

	_, err2 := rw.statement.Exec(values...)
	if err2 != nil {
		pqErr := err2.(*pq.Error)
		if pqErr.Code == "23505" { //duplicate
			rw.ds.logger.Trace(fmt.Sprintf("duplicate key %v (%s) for %s", row[rw.table.PrimaryKey()], rw.table.PrimaryKey(), rw.table.Name()))
			// TODO update
		} else {
			return &push.Error{Description: err2.Error()}
		}
	}

	return nil
}

func (rw *PostgresRowWriter) truncate() *push.Error {
	stm := "TRUNCATE TABLE " + rw.table.Name() + " CASCADE"
	rw.ds.logger.Debug(stm)
	_, err := rw.ds.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

func (rw *PostgresRowWriter) disableConstraints() *push.Error {
	stm := "ALTER TABLE " + rw.table.Name() + " DISABLE TRIGGER ALL"
	rw.ds.logger.Debug(stm)
	_, err := rw.ds.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

func (rw *PostgresRowWriter) enableConstraints() *push.Error {
	stm := "ALTER TABLE " + rw.table.Name() + " ENABLE TRIGGER ALL"
	rw.ds.logger.Debug(stm)
	_, err := rw.ds.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}
