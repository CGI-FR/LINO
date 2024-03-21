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

package pull

import (
	"github.com/cgi-fr/lino/internal/infra/commonsql"
	"github.com/cgi-fr/lino/pkg/pull"

	_ "github.com/ziutek/mymysql/godrv"
)

// MariadbDataSourceFactory exposes methods to create new Mariadb pullers.
type MariadbDataSourceFactory struct{}

// NewMariadbDataSourceFactory creates a new mariadb datasource factory.
func NewMariadbDataSourceFactory() *MariadbDataSourceFactory {
	return &MariadbDataSourceFactory{}
}

// New return a Mariadb puller
func (e *MariadbDataSourceFactory) New(url string, schema string, options ...pull.DataSourceOption) pull.DataSource {
	ds := &SQLDataSource{
		url:     url,
		schema:  schema,
		dialect: commonsql.MariadbDialect{},
	}

	for _, option := range options {
		option(ds)
	}

	return ds
}
