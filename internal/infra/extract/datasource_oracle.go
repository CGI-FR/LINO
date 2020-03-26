package extract

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/extract"

	// import oracle connector
	_ "github.com/lib/pq"
)

// OracleDataSourceFactory exposes methods to create new Oracle extractors.
type OracleDataSourceFactory struct {
	logger extract.Logger
}

// NewOracleDataSourceFactory creates a new oracle datasource factory.
func NewOracleDataSourceFactory(l extract.Logger) *OracleDataSourceFactory {
	return &OracleDataSourceFactory{l}
}

// New return a Oracle extractor
func (e *OracleDataSourceFactory) New(url string) extract.DataSource {
	return NewOracleDataSource(url, e.logger)
}

// OracleDataSource read data from a Oracle database.
type OracleDataSource struct {
	url    string
	logger extract.Logger
	dbx    *sqlx.DB
	db     *sql.DB
}

// NewOracleDataSource creates a new oracle datasource.
func NewOracleDataSource(url string, logger extract.Logger) *OracleDataSource {
	return &OracleDataSource{
		url:    url,
		logger: logger,
	}
}

// Open a connection to the oracle DB
func (ds *OracleDataSource) Open() *extract.Error {
	db, err := dburl.Open(ds.url)
	if err != nil {
		return &extract.Error{Description: err.Error()}
	}

	ds.db = db

	u, err := dburl.Parse(ds.url)
	if err != nil {
		return &extract.Error{Description: err.Error()}
	}

	ds.dbx = sqlx.NewDb(db, u.Unaliased)

	err = ds.dbx.Ping()
	if err != nil {
		return &extract.Error{Description: err.Error()}
	}

	return nil
}

// Close a connection to the oracle DB
func (ds *OracleDataSource) Close() *extract.Error {
	err := ds.dbx.Close()
	if err != nil {
		return &extract.Error{Description: err.Error()}
	}
	return nil
}

// RowReader iterate over rows in table with filter
func (ds *OracleDataSource) RowReader(source extract.Table, filter extract.Filter) (extract.RowReader, *extract.Error) {
	sql := &strings.Builder{}
	sql.Write([]byte("SELECT * FROM "))
	sql.Write([]byte(source.Name()))
	sql.Write([]byte(" "))
	if len(filter.Values()) > 0 || filter.Limit() > 0 {
		sql.Write([]byte("WHERE "))
	}
	values := []interface{}{}
	for key, value := range filter.Values() {
		sql.Write([]byte(key))
		values = append(values, value)
		sql.Write([]byte("=:v"))
		fmt.Fprintf(sql, "%v", len(values))
		sql.Write([]byte(" AND "))
	}
	if filter.Limit() > 0 {
		fmt.Fprintf(sql, "rownum <= %v AND ", filter.Limit())
	}
	if len(filter.Values()) > 0 || filter.Limit() > 0 {
		sql.Write([]byte("1=1 "))
	}

	if ds.logger != nil {
		printSQL := sql.String()
		ds.logger.Debug(fmt.Sprint(printSQL))
		printSQL = strings.TrimSuffix(printSQL, " AND 1=1")
		for i, v := range values {
			printSQL = strings.ReplaceAll(printSQL, fmt.Sprintf(":v%v", i+1), fmt.Sprintf("%v", v))
		}
		ds.logger.Debug(fmt.Sprint(printSQL))
	}

	rows, err := ds.dbx.Queryx(sql.String(), values...)
	if err != nil {
		return nil, &extract.Error{Description: err.Error()}
	}

	return &OracleDataIterator{rows}, nil
}

// OracleDataIterator read data from a Oracle database.
type OracleDataIterator struct {
	rows *sqlx.Rows
}

// Next reads the next rows if it exists.
func (di *OracleDataIterator) Next() bool {
	if di.rows == nil {
		return false
	}
	return di.rows.Next()
}

// Value returns the last read row.
func (di *OracleDataIterator) Value() (extract.Row, *extract.Error) {
	columns, err := di.rows.Columns()
	if err != nil {
		return nil, &extract.Error{Description: err.Error()}
	}

	values, err := di.rows.SliceScan()
	if err != nil {
		return nil, &extract.Error{Description: err.Error()}
	}

	row := extract.Row{}
	for i, column := range columns {
		row[column] = values[i]
	}

	return row, nil
}
