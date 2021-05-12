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

package push

import (
	"fmt"
	"strings"

	"github.com/cgi-fr/lino/pkg/push"
	"github.com/lib/pq"
)

// PostgresDataDestinationFactory exposes methods to create new Postgres pullers.
type PostgresDataDestinationFactory struct {
	logger push.Logger
}

// NewPostgresDataDestinationFactory creates a new postgres datadestination factory.
func NewPostgresDataDestinationFactory(l push.Logger) *PostgresDataDestinationFactory {
	return &PostgresDataDestinationFactory{l}
}

// New return a Postgres pusher
func (e *PostgresDataDestinationFactory) New(url string, schema string) push.DataDestination {
	return NewSQLDataDestination(url, schema, PostgresDialect{}, e.logger)
}

// PostgresDialect inject postgres variations
type PostgresDialect struct{}

// Placeholde return the variable format for postgres
func (d PostgresDialect) Placeholder(position int) string {
	return fmt.Sprintf("$%d", position)
}

// EnableConstraintsStatement generate statments to activate constraintes
func (d PostgresDialect) EnableConstraintsStatement(tableName string) string {
	return fmt.Sprintf("ALTER TABLE %s ENABLE TRIGGER ALL", tableName)
}

// DisableConstraintsStatement generate statments to deactivate constraintes
func (d PostgresDialect) DisableConstraintsStatement(tableName string) string {
	return fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER ALL", tableName)
}

// TruncateStatement generate statement to truncat table content
func (d PostgresDialect) TruncateStatement(tableName string) string {
	return fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName)
}

// InsertStatement  generate insert statement
func (d PostgresDialect) InsertStatement(tableName string, columns []string, values []string, primaryKeys []string) string {
	protectedColumns := []string{}
	for _, c := range columns {
		protectedColumns = append(protectedColumns, fmt.Sprintf("\"%s\"", c))
	}
	if len(primaryKeys) > 0 {
		return fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) ON CONFLICT (%s) DO NOTHING", tableName, strings.Join(protectedColumns, ","), strings.Join(values, ","), strings.Join(primaryKeys, ","))
	}
	return fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", tableName, strings.Join(protectedColumns, ","), strings.Join(values, ","))
}

func (d PostgresDialect) UpdateStatement(tableName string, columns []string, uValues []string, primaryKeys []string, pValues []string) (string, *push.Error) {
	sql := &strings.Builder{}
	sql.Write([]byte("UPDATE "))
	sql.Write([]byte(tableName))
	sql.Write([]byte(" SET "))
	for index, column := range columns {
		sql.Write([]byte(column))
		fmt.Fprint(sql, "=")
		fmt.Fprint(sql, uValues[index])
		if index+1 < len(columns) {
			sql.Write([]byte(", "))
		}
	}
	if len(primaryKeys) > 0 {
		sql.Write([]byte(" WHERE "))
	} else {
		return "", &push.Error{Description: fmt.Sprintf("can't update table [%s] because no primary key is defined", tableName)}
	}
	for index, pk := range primaryKeys {
		sql.Write([]byte(pk))
		fmt.Fprint(sql, "=")
		fmt.Fprint(sql, pValues[index])
		if index+1 < len(primaryKeys) {
			sql.Write([]byte(" AND "))
		}
	}
	return sql.String(), nil
}

// IsDuplicateError check if error is a duplicate error
func (d PostgresDialect) IsDuplicateError(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "23505"
}

// ConvertValue before load
func (d PostgresDialect) ConvertValue(from push.Value) push.Value {
	return from
}
