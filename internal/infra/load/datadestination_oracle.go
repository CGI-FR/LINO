package load

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/load"

	// import Oracle connector
	_ "github.com/godror/godror"
)

// OracleDataDestinationFactory exposes methods to create new Oracle extractors.
type OracleDataDestinationFactory struct {
	logger load.Logger
}

// NewOracleDataDestinationFactory creates a new Oracle datadestination factory.
func NewOracleDataDestinationFactory(l load.Logger) *OracleDataDestinationFactory {
	return &OracleDataDestinationFactory{l}
}

// New return a Oracle loader
func (e *OracleDataDestinationFactory) New(url string) load.DataDestination {
	return NewOracleDataDestination(url, e.logger)
}

// OracleDataDestination read data from a Oracle database.
type OracleDataDestination struct {
	url       string
	logger    load.Logger
	db        *sqlx.DB
	rowWriter map[string]*OracleRowWriter
	mode      load.Mode
}

// NewOracleDataDestination creates a new Oracle datadestination.
func NewOracleDataDestination(url string, logger load.Logger) *OracleDataDestination {
	return &OracleDataDestination{
		url:       url,
		logger:    logger,
		rowWriter: map[string]*OracleRowWriter{},
	}
}

// Close Oracle connections
func (ds *OracleDataDestination) Close() *load.Error {
	for _, rw := range ds.rowWriter {
		rw.close()
	}

	err := ds.db.Close()
	if err != nil {
		return &load.Error{Description: err.Error()}
	}
	return nil
}

// Open Oracle Connections
func (ds *OracleDataDestination) Open(plan load.Plan, mode load.Mode) *load.Error {
	ds.mode = mode

	ds.logger.Info(fmt.Sprintf("connecting to %s...", ds.url))
	db, err := dburl.Open(ds.url)
	if err != nil {
		return &load.Error{Description: err.Error()}
	}

	u, err := dburl.Parse(ds.url)
	if err != nil {
		return &load.Error{Description: err.Error()}
	}

	dbx := sqlx.NewDb(db, u.Unaliased)

	err = dbx.Ping()
	if err != nil {
		return &load.Error{Description: err.Error()}
	}

	ds.db = dbx

	for _, table := range plan.Tables() {
		rw := NewOracleRowWriter(table, ds)
		err := rw.open()
		if err != nil {
			return err
		}

		ds.rowWriter[table.Name()] = rw
	}

	return nil
}

// RowWriter return Oracle table writer
func (ds *OracleDataDestination) RowWriter(table load.Table) (load.RowWriter, *load.Error) {
	rw, ok := ds.rowWriter[table.Name()]
	if ok {
		return rw, nil
	}

	rw = NewOracleRowWriter(table, ds) //TODO
	err := rw.open()
	if err != nil {
		return nil, err
	}

	ds.rowWriter[table.Name()] = rw
	return rw, nil
}

// OracleRowWriter write data to a Oracle table.
type OracleRowWriter struct {
	table              load.Table
	ds                 *OracleDataDestination
	duplicateKeysCache map[load.Value]struct{}
	statement          *sql.Stmt
	headers            []string
}

// NewOracleRowWriter creates a new Oracle row writer.
func NewOracleRowWriter(table load.Table, ds *OracleDataDestination) *OracleRowWriter {
	return &OracleRowWriter{
		table: table,
		ds:    ds,
	}
}

// open table writer
func (rw *OracleRowWriter) open() *load.Error {
	rw.ds.logger.Debug(fmt.Sprintf("open table with mode %s", rw.ds.mode))
	if rw.ds.mode == load.Truncate {
		err := rw.truncate()
		if err != nil {
			return &load.Error{Description: err.Error()}
		}
	}

	err2 := rw.disableConstraints()
	if err2 != nil {
		return &load.Error{Description: err2.Error()}
	}
	rw.duplicateKeysCache = map[load.Value]struct{}{}
	return nil
}

// close table writer
func (rw *OracleRowWriter) close() *load.Error {
	if rw.statement != nil {
		err := rw.statement.Close()
		if err != nil {
			return &load.Error{Description: err.Error()}
		}
	}

	return rw.enableConstraints()
}

func (rw *OracleRowWriter) createStatement(row load.Row) *load.Error {
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

	/* #nosec */
	prepareStmt := "INSERT INTO " + rw.table.Name() + "(" + strings.Join(names, ",") + ") VALUES(" + strings.Join(valuesVar, ",") + ")"
	rw.ds.logger.Debug(prepareStmt)
	// TODO: Create an update statement

	stmt, err := rw.ds.db.Prepare(prepareStmt)
	if err != nil {
		return &load.Error{Description: err.Error()}
	}
	rw.statement = stmt
	rw.headers = names
	return nil
}

// Write
func (rw *OracleRowWriter) Write(row load.Row) *load.Error {
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
		return &load.Error{Description: err2.Error()}
	}

	return nil
}

func (rw *OracleRowWriter) truncate() *load.Error {
	stm := "TRUNCATE TABLE " + rw.table.Name() + " CASCADE"
	rw.ds.logger.Debug(stm)
	_, err := rw.ds.db.Exec(stm)
	if err != nil {
		return &load.Error{Description: err.Error()}
	}
	return nil
}

func (rw *OracleRowWriter) disableConstraints() *load.Error {
	stm := "ALTER TABLE " + rw.table.Name() + " DISABLE TRIGGER ALL"
	rw.ds.logger.Debug(stm)
	_, err := rw.ds.db.Exec(stm)
	if err != nil {
		return &load.Error{Description: err.Error()}
	}
	return nil
}

func (rw *OracleRowWriter) enableConstraints() *load.Error {
	stm := "ALTER TABLE " + rw.table.Name() + " ENABLE TRIGGER ALL"
	rw.ds.logger.Debug(stm)
	_, err := rw.ds.db.Exec(stm)
	if err != nil {
		return &load.Error{Description: err.Error()}
	}
	return nil
}
