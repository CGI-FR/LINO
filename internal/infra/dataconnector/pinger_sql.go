package dataconnector

import (
	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
)

type SQLDataPingerFactory struct {
	logger dataconnector.Logger
}

// NewSQLDataPinger creates a new SQL pinger.
func NewSQLDataPingerFactory(logger dataconnector.Logger) *SQLDataPingerFactory {
	return &SQLDataPingerFactory{
		logger: logger,
	}
}

func (pdpf SQLDataPingerFactory) New(url string) dataconnector.DataPinger {
	return NewSQLDataPinger(url, pdpf.logger)
}

func NewSQLDataPinger(url string, logger dataconnector.Logger) SQLDataPinger {
	return SQLDataPinger{url, logger}
}

type SQLDataPinger struct {
	url    string
	logger dataconnector.Logger
}

func (pdp SQLDataPinger) Ping() *dataconnector.Error {
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
