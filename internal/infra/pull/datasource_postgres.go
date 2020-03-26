package pull

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/pull"

	// import postgresql connector
	_ "github.com/lib/pq"
)

// PostgresDataSourceFactory exposes methods to create new Postgres pullers.
type PostgresDataSourceFactory struct {
	logger pull.Logger
}

// NewPostgresDataSourceFactory creates a new postgres datasource factory.
func NewPostgresDataSourceFactory(l pull.Logger) *PostgresDataSourceFactory {
	return &PostgresDataSourceFactory{l}
}

// New return a Postgres puller
func (e *PostgresDataSourceFactory) New(url string) pull.DataSource {
	return NewPostgresDataSource(url, e.logger)
}

// PostgresDataSource read data from a PostgreSQL database.
type PostgresDataSource struct {
	url    string
	logger pull.Logger
	dbx    *sqlx.DB
	db     *sql.DB
}

// NewPostgresDataSource creates a new postgres datasource.
func NewPostgresDataSource(url string, logger pull.Logger) *PostgresDataSource {
	return &PostgresDataSource{
		url:    url,
		logger: logger,
	}
}

// Open a connection to the postgres DB
func (ds *PostgresDataSource) Open() *pull.Error {
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

// Close a connection to the postgres DB
func (ds *PostgresDataSource) Close() *pull.Error {
	err := ds.dbx.Close()
	if err != nil {
		return &pull.Error{Description: err.Error()}
	}
	return nil
}

// RowReader iterate over rows in table with filter
func (ds *PostgresDataSource) RowReader(source pull.Table, filter pull.Filter) (pull.RowReader, *pull.Error) {
	sql := &strings.Builder{}
	sql.Write([]byte("SELECT * FROM "))
	sql.Write([]byte(source.Name()))
	sql.Write([]byte(" "))
	if len(filter.Values()) > 0 {
		sql.Write([]byte("WHERE "))
	}
	values := []interface{}{}
	for key, value := range filter.Values() {
		sql.Write([]byte(key))
		values = append(values, value)
		sql.Write([]byte("=$"))
		fmt.Fprintf(sql, "%v", len(values))
		sql.Write([]byte(" AND "))
	}
	if len(filter.Values()) > 0 {
		sql.Write([]byte("1=1 "))
	}

	if filter.Limit() > 0 {
		fmt.Fprintf(sql, "LIMIT %v", filter.Limit())
	}

	if ds.logger != nil {
		printSQL := sql.String()
		printSQL = strings.TrimSuffix(printSQL, " AND 1=1")
		for i, v := range values {
			printSQL = strings.ReplaceAll(printSQL, fmt.Sprintf("$%v", i+1), fmt.Sprintf("%v", v))
		}
		ds.logger.Debug(fmt.Sprint(printSQL))
	}

	rows, err := ds.dbx.Queryx(sql.String(), values...)
	if err != nil {
		return nil, &pull.Error{Description: err.Error()}
	}

	return &PostgresDataIterator{rows}, nil
}

// PostgresDataIterator read data from a PostgreSQL database.
type PostgresDataIterator struct {
	rows *sqlx.Rows
}

// Next reads the next rows if it exists.
func (di *PostgresDataIterator) Next() bool {
	if di.rows == nil {
		return false
	}
	return di.rows.Next()
}

// Value returns the last read row.
func (di *PostgresDataIterator) Value() (pull.Row, *pull.Error) {
	columns, err := di.rows.Columns()
	if err != nil {
		return nil, &pull.Error{Description: err.Error()}
	}

	values, err := di.rows.SliceScan()
	if err != nil {
		return nil, &pull.Error{Description: err.Error()}
	}

	row := pull.Row{}
	for i, column := range columns {
		row[column] = values[i]
	}

	return row, nil
}
