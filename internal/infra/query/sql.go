package query

import (
	"fmt"
	"time"

	"github.com/cgi-fr/lino/pkg/query"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/xo/dburl"
)

type DataReader struct {
	rows  *sqlx.Rows
	value any
	err   error
}

func (dr *DataReader) Next() bool {
	if dr.rows == nil {
		return false
	}
	if dr.rows.Next() {
		columns, err := dr.rows.Columns()
		if err != nil {
			dr.err = err
			return false
		}

		values, err := dr.rows.SliceScan()
		if err != nil {
			dr.err = err
			return false
		}

		row := map[string]any{}
		for i, column := range columns {
			row[column] = values[i]
		}
		dr.value = row
		return true
	}
	if dr.rows.Err() != nil {
		dr.err = dr.rows.Err()
	}
	return false
}

func (dr *DataReader) Value() any {
	return dr.value
}

func (dr *DataReader) Error() error {
	return dr.err
}

type DataSource struct {
	url         string
	dbx         *sqlx.DB
	maxLifeTime time.Duration
}

func (ds *DataSource) Open() error {
	u, err := dburl.Parse(ds.url)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	db, err := dburl.Open(ds.url)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	ds.dbx = sqlx.NewDb(db, u.UnaliasedDriver)

	ds.dbx.SetConnMaxLifetime(ds.maxLifeTime)
	ds.dbx.SetConnMaxIdleTime(ds.maxLifeTime)

	err = ds.dbx.Ping()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (ds *DataSource) Close() error {
	if err := ds.dbx.Close(); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (ds *DataSource) Query(query string) (query.DataReader, error) {
	rows, err := ds.dbx.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	log.Info().Str("query", query).Msg("success executing SQL query")

	return &DataReader{rows, nil, nil}, nil
}

type DataSourceFactory struct{}

func (dsf DataSourceFactory) New(url string, maxLifeTime time.Duration) query.DataSource {
	return &DataSource{
		url:         url,
		dbx:         nil,
		maxLifeTime: maxLifeTime,
	}
}
