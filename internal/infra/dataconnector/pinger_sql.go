// Copyright (C) 2021 CGI France
//
// This file is part of LINO.
//
// LINO is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// LINO is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with LINO.  If not, see <http://www.gnu.org/licenses/>.

package dataconnector

import (
	"github.com/cgi-fr/lino/pkg/dataconnector"
	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
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
