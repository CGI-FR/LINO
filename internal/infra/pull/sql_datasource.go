package pull

import (
	"database/sql"
	"fmt"
	"strings"

	"makeit.imfr.cgi.com/lino/pkg/pull"

	// import oracle connector
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/xo/dburl"
)

// DataSource to read in the pull process.
type SQLDataSource struct {
	url     string
	logger  pull.Logger
	dbx     *sqlx.DB
	db      *sql.DB
	dialect SQLDialect
}

// Open a connection to the SQL DB
func (ds *SQLDataSource) Open() *pull.Error {
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

// RowReader iterate over rows in table with filter
func (ds *SQLDataSource) RowReader(source pull.Table, filter pull.Filter) (pull.RowReader, *pull.Error) {
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
		fmt.Fprint(sql, "=")
		fmt.Fprint(sql, ds.dialect.Placeholder(len(values)))
		if len(values) < len(filter.Values()) {
			sql.Write([]byte(" AND "))
		}
	}

	if filter.Limit() > 0 {
		fmt.Fprint(sql, ds.dialect.Limit(filter.Limit()))
	}

	if ds.logger != nil {
		printSQL := sql.String()
		for i, v := range values {
			printSQL = strings.ReplaceAll(printSQL, ds.dialect.Placeholder(i+1), fmt.Sprintf("%v", v))
		}
		ds.logger.Debug(fmt.Sprint(printSQL))
	}

	rows, err := ds.dbx.Queryx(sql.String(), values...)
	if err != nil {
		return nil, &pull.Error{Description: err.Error()}
	}

	return &SQLDataIterator{rows}, nil
}

// Close a connection to the SQL DB
func (ds *SQLDataSource) Close() *pull.Error {
	err := ds.dbx.Close()
	if err != nil {
		return &pull.Error{Description: err.Error()}
	}
	return nil
}

// SQLDataIterator read data from a SQL database.
type SQLDataIterator struct {
	rows *sqlx.Rows
}

// Next reads the next rows if it exists.
func (di *SQLDataIterator) Next() bool {
	if di.rows == nil {
		return false
	}
	return di.rows.Next()
}

// Value returns the last read row.
func (di *SQLDataIterator) Value() (pull.Row, *pull.Error) {
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

// SQLDialect to inject SQL variations
type SQLDialect interface {

	// Placeholder format variable in query
	Placeholder(int) string
	// Limit format limitation clause
	Limit(uint) string
}
