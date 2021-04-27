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

	// import postgresql connector
	_ "github.com/lib/pq"
)

// PostgresDataSourceFactory exposes methods to create new Postgres pullers.
type PostgresDataSourceFactory struct {
	logger pull.Logger
}

// NewPostgresDataSourceFactory creates a new postgres datasource factory.
func NewPostgresDataSourceFactory(l pull.Logger) *PostgresDataSourceFactory {
	return &PostgresDataSourceFactory{l}
}

// New return a Postgres puller
func (e *PostgresDataSourceFactory) New(url string, schema string) pull.DataSource {
	return &SQLDataSource{
		url:     url,
		schema:  schema,
		dialect: PostgresDialect{},
		logger:  e.logger,
	}
}

// PostgresDialect implement postgres SQL variations
type PostgresDialect struct{}

func (pd PostgresDialect) Placeholder(position int) string {
	return fmt.Sprintf("$%d", position)
}

func (pd PostgresDialect) Limit(limit uint) string {
	return fmt.Sprintf(" LIMIT %d", limit)
}
