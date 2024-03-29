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

//go:build db2
// +build db2

package pull

import (
	_ "github.com/ibmdb/go_ibm_db"

	"github.com/cgi-fr/lino/internal/infra/commonsql"
	"github.com/cgi-fr/lino/pkg/pull"
)

// Db2DataSourceFactory exposes methods to create new Db2 pullers.
type Db2DataSourceFactory struct{}

// NewDb2DataSourceFactory creates a new oracle datasource factory.
func NewDb2DataSourceFactory() *Db2DataSourceFactory {
	return &Db2DataSourceFactory{}
}

// New return a Db2 puller
func (e *Db2DataSourceFactory) New(url string, schema string) pull.DataSource {
	return &SQLDataSource{
		url:     url,
		schema:  schema,
		dialect: commonsql.Db2Dialect{},
	}
}
