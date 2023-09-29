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
	"fmt"

	"github.com/cgi-fr/lino/pkg/pull"

	// import SQLServersql connector
	_ "github.com/microsoft/go-mssqldb"
)

// SQLServerDataSourceFactory exposes methods to create new SQLServer pullers.
type SQLServerDataSourceFactory struct{}

// NewSQLServerDataSourceFactory creates a new SQLServer datasource factory.
func NewSQLServerDataSourceFactory() *SQLServerDataSourceFactory {
	return &SQLServerDataSourceFactory{}
}

// New return a SQLServer puller
func (e *SQLServerDataSourceFactory) New(url string, schema string) pull.DataSource {
	return &SQLDataSource{
		url:     url,
		schema:  schema,
		dialect: SQLServerDialect{},
	}
}

// SQLServerDialect implement SQLServer SQL variations

type SQLServerDialect struct{}

func (sd SQLServerDialect) Placeholder(position int) string {
	return fmt.Sprintf("@p%d", position)
}

func (sd SQLServerDialect) Limit(limit uint) string {
	return fmt.Sprintf("TOP %d", limit)
}
