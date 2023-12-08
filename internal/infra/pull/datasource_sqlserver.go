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

// Limit method is adjusted to be compatible with SQL Server
func (sd SQLServerDialect) Limit(limit uint) string {
	return fmt.Sprintf(" TOP %d", limit)
}

// CreateSelect generate a SQL request in the correct order
func (sd SQLServerDialect) CreateSelect(sel string, where string, limit string, columns string, from string) string {
	return fmt.Sprintf("%s %s %s %s %s", sel, limit, columns, from, where)
}