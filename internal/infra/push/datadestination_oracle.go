package push

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/push"

	// import Oracle connector
	_ "github.com/godror/godror"
)

// OracleDataDestinationFactory exposes methods to create new Oracle extractors.
type OracleDataDestinationFactory struct {
	logger push.Logger
}

// NewOracleDataDestinationFactory creates a new Oracle datadestination factory.
func NewOracleDataDestinationFactory(l push.Logger) *OracleDataDestinationFactory {
	return &OracleDataDestinationFactory{l}
}

// New return a Oracle pusher
func (e *OracleDataDestinationFactory) New(url string) push.DataDestination {
	return NewOracleDataDestination(url, e.logger)
}

// OracleDataDestination read data from a Oracle database.
type OracleDataDestination struct {
	url       string
	logger    push.Logger
	db        *sqlx.DB
	rowWriter map[string]*OracleRowWriter
	mode      push.Mode
}

// NewOracleDataDestination creates a new Oracle datadestination.
func NewOracleDataDestination(url string, logger push.Logger) *OracleDataDestination {
	return &OracleDataDestination{
		url:       url,
		logger:    logger,
		rowWriter: map[string]*OracleRowWriter{},
	}
}

// Close Oracle connections
func (ds *OracleDataDestination) Close() *push.Error {
	for _, rw := range ds.rowWriter {
		rw.close()
	}

	err := ds.db.Close()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

// Open Oracle Connections
func (ds *OracleDataDestination) Open(plan push.Plan, mode push.Mode) *push.Error {
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
func (ds *OracleDataDestination) RowWriter(table push.Table) (push.RowWriter, *push.Error) {
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
	table              push.Table
	ds                 *OracleDataDestination
	duplicateKeysCache map[push.Value]struct{}
	statement          *sql.Stmt
	headers            []string
}

// NewOracleRowWriter creates a new Oracle row writer.
func NewOracleRowWriter(table push.Table, ds *OracleDataDestination) *OracleRowWriter {
	return &OracleRowWriter{
		table: table,
		ds:    ds,
	}
}

// open table writer
func (rw *OracleRowWriter) open() *push.Error {
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
func (rw *OracleRowWriter) close() *push.Error {
	if rw.statement != nil {
		err := rw.statement.Close()
		if err != nil {
			return &push.Error{Description: err.Error()}
		}
	}

	return rw.enableConstraints()
}

func (rw *OracleRowWriter) createStatement(row push.Row) *push.Error {
	if rw.statement != nil {
		return nil
	}

	names := []string{}
	valuesVar := []string{}

	i := 1
	for k := range row {
		names = append(names, k)
		valuesVar = append(valuesVar, fmt.Sprintf(":v%d", i))
		i++
	}

	/* #nosec */
	prepareStmt := "INSERT INTO " + rw.table.Name() + "(" + strings.Join(names, ",") + ") VALUES(" + strings.Join(valuesVar, ",") + ")"
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

func (rw *OracleRowWriter) addToCache(row push.Row) {
	rw.duplicateKeysCache[rw.computeRowSum(row)] = struct{}{}
}

func (rw *OracleRowWriter) isInCache(row push.Row) bool {
	_, ok := rw.duplicateKeysCache[rw.computeRowSum(row)]
	return ok
}

func (rw *OracleRowWriter) computeRowSum(row push.Row) string {
	sum := ""
	for _, pk := range rw.table.PrimaryKey() {
		sum = fmt.Sprintf("%s|,-%v", sum, row[pk])
	}
	return sum
}

// Write
func (rw *OracleRowWriter) Write(row push.Row) *push.Error {
	if ok := rw.isInCache(row); ok {
		rw.ds.logger.Trace(fmt.Sprintf("duplicate key in dataset %v (%s) for %s", row, rw.table.PrimaryKey(), rw.table.Name()))
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
	rw.ds.logger.Trace(fmt.Sprint(values))

	_, err2 := rw.statement.Exec(values...)
	if err2 != nil {
		return &push.Error{Description: err2.Error()}
	}

	return nil
}

func (rw *OracleRowWriter) truncate() *push.Error {
	stm := "TRUNCATE TABLE " + rw.table.Name() + " CASCADE"
	rw.ds.logger.Debug(stm)
	_, err := rw.ds.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

func (rw *OracleRowWriter) disableConstraints() *push.Error {
	/* #nosec */
	stm := fmt.Sprintf(`BEGIN
	FOR c IN
	(SELECT c.owner, c.table_name, c.constraint_name
	 FROM user_constraints c, user_tables t
	 WHERE c.table_name = t.table_name
	 AND c.table_name = '%s'
	 AND c.status = 'ENABLED'
	 AND NOT (t.iot_type IS NOT NULL AND c.constraint_type = 'P')
	 ORDER BY c.constraint_type DESC)
	LOOP
	  dbms_utility.exec_ddl_statement('alter table "' || c.owner || '"."' || c.table_name || '" disable constraint ' || c.constraint_name);
	END LOOP;
  END;
  `, rw.table.Name())

	rw.ds.logger.Debug(stm)
	_, err := rw.ds.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

func (rw *OracleRowWriter) enableConstraints() *push.Error {
	/* #nosec */
	stm := fmt.Sprintf(`BEGIN
	FOR c IN
	(SELECT c.owner, c.table_name, c.constraint_name
	 FROM user_constraints c, user_tables t
	 WHERE c.table_name = t.table_name
	 AND c.table_name = '%s'
	 AND c.status = 'DISABLED'
	 ORDER BY c.constraint_type)
	LOOP
	  dbms_utility.exec_ddl_statement('alter table "' || c.owner || '"."' || c.table_name || '" enable constraint ' || c.constraint_name);
	END LOOP;
  END;
  `, rw.table.Name())

	rw.ds.logger.Debug(stm)
	_, err := rw.ds.db.Exec(stm)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}
