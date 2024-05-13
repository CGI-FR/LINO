package query

import (
	"fmt"

	"github.com/cgi-fr/lino/pkg/query"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/xo/dburl"
)

type DataReader struct{}

func (dr *DataReader) Next() bool {
	return false
}

func (dr *DataReader) Value() any {
	return nil
}

func (dr *DataReader) Error() error {
	return nil
}

type DataSource struct {
	url string
	dbx *sqlx.DB
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
	if _, err := ds.dbx.Exec(query); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	log.Info().Str("query", query).Msg("success executing SQL query")

	return nil, nil
}

type DataSourceFactory struct{}

func (dsf DataSourceFactory) New(url string) query.DataSource {
	return &DataSource{
		url: url,
		dbx: nil,
	}
}
