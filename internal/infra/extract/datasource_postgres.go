package extract

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/extract"

	// import postgresql connector
	_ "github.com/lib/pq"
)

// PostgresDataSourceFactory exposes methods to create new Postgres extractors.
type PostgresDataSourceFactory struct {
	logger extract.Logger
}

// NewPostgresDataSourceFactory creates a new postgres datasource factory.
func NewPostgresDataSourceFactory(l extract.Logger) *PostgresDataSourceFactory {
	return &PostgresDataSourceFactory{l}
}

// New return a Postgres extractor
func (e *PostgresDataSourceFactory) New(url string) extract.DataSource {
	return NewPostgresDataSource(url, e.logger)
}

// PostgresDataSource read data from a PostgreSQL database.
type PostgresDataSource struct {
	url    string
	logger extract.Logger
}

// NewPostgresDataSource creates a new postgres datasource.
func NewPostgresDataSource(url string, logger extract.Logger) *PostgresDataSource {
	return &PostgresDataSource{
		url:    url,
		logger: logger,
	}
}

func (ds *PostgresDataSource) Read(source extract.Table, filter extract.Filter) (extract.DataIterator, *extract.Error) {
	db, err := dburl.Open(ds.url)
	if err != nil {
		return nil, &extract.Error{Description: err.Error()}
	}
	defer db.Close()

	u, err := dburl.Parse(ds.url)
	if err != nil {
		return nil, &extract.Error{Description: err.Error()}
	}

	dbx := sqlx.NewDb(db, u.Unaliased)

	err = dbx.Ping()
	if err != nil {
		return nil, &extract.Error{Description: err.Error()}
	}

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

	rows, err := dbx.Queryx(sql.String(), values...)
	if err != nil {
		return nil, &extract.Error{Description: err.Error()}
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
func (di *PostgresDataIterator) Value() (extract.Row, *extract.Error) {
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
