package dataconnector

import (
	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
	// import postgresql connector
)

type PostgresDataPingerFactory struct {
	logger dataconnector.Logger
}

// NewPostgresDataPinger creates a new postgres pinger.
func NewPostgresDataPingerFactory(logger dataconnector.Logger) *PostgresDataPingerFactory {
	return &PostgresDataPingerFactory{
		logger: logger,
	}
}

func (pdpf PostgresDataPingerFactory) New(url string) dataconnector.DataPinger {
	return NewPostgresDataPinger(url, pdpf.logger)
}

func NewPostgresDataPinger(url string, logger dataconnector.Logger) PostgresDataPinger {
	return PostgresDataPinger{url, logger}
}

type PostgresDataPinger struct {
	url    string
	logger dataconnector.Logger
}

func (pdp PostgresDataPinger) Ping() *dataconnector.Error {
	db, err := dburl.Open(pdp.url)
	if err != nil {
		return &dataconnector.Error{Description: err.Error()}
	}

	u, err := dburl.Parse(pdp.url)
	if err != nil {
		return &dataconnector.Error{Description: err.Error()}
	}

	dbx := sqlx.NewDb(db, u.Unaliased)

	err = dbx.Ping()
	if err != nil {
		return &dataconnector.Error{Description: err.Error()}
	}

	return nil
}
