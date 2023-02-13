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

// MariadbDataDestinationFactory exposes methods to create new Mariadb pullers.
type MariadbDataDestinationFactory struct{}

// NewMariadbDataDestinationFactory creates a new mariadb datadestination factory.
func NewMariadbDataDestinationFactory() *MariadbDataDestinationFactory {
	return &MariadbDataDestinationFactory{}
}

// New return a Mariadb pusher
func (e *MariadbDataDestinationFactory) New(url string, schema string) push.DataDestination {
	return NewSQLDataDestination(url, schema, MariadbDialect{})
}

// MariadbDialect inject mariadb variations
type MariadbDialect struct{}

// Placeholde return the variable format for mariadb
func (d MariadbDialect) Placeholder(position int) string {
	return "?"
}

// EnableConstraintsStatement generate statments to activate constraintes
func (d MariadbDialect) EnableConstraintsStatement(tableName string) string {
	return "SET GLOBAL FOREIGN_KEY_CHECKS=1"
}

// DisableConstraintsStatement generate statments to deactivate constraintes
func (d MariadbDialect) DisableConstraintsStatement(tableName string) string {
	return "SET GLOBAL FOREIGN_KEY_CHECKS=0"
}

// TruncateStatement generate statement to truncat table content (ON DELETE CASCADE must be set to TRUNCATE child tables)
func (d MariadbDialect) TruncateStatement(tableName string) string {
	return fmt.Sprintf("TRUNCATE TABLE %s", tableName)
}

// InsertStatement  generate insert statement
func (d MariadbDialect) InsertStatement(tableName string, columns []string, values []string, primaryKeys []string) string {
	protectedColumns := []string{}
	for _, c := range columns {
		protectedColumns = append(protectedColumns, fmt.Sprintf("`%s`", c))
	}
	if len(primaryKeys) > 0 {
		sql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", tableName, strings.Join(protectedColumns, ","), strings.Join(values, ","))
		ret := strings.TrimSuffix(sql, ",")
		return ret
	}
	return fmt.Sprintf("INSERT IGNORE INTO %s(%s) VALUES(%s)", tableName, strings.Join(protectedColumns, ","), strings.Join(values, ","))
}

func (d MariadbDialect) UpdateStatement(tableName string, columns []string, uValues []string, primaryKeys []string, pValues []string, where push.Row) (string, []string, *push.Error) {
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
		return "", []string{}, &push.Error{Description: fmt.Sprintf("can't update table [%s] because no primary key is defined", tableName)}
	}
	for index, pk := range primaryKeys {
		sql.Write([]byte(pk))
		fmt.Fprint(sql, "=")
		fmt.Fprint(sql, pValues[index])
		if index+1 < len(primaryKeys) {
			sql.Write([]byte(" AND "))
		}
	}
	return sql.String(), append(columns, primaryKeys...), nil
}

// IsDuplicateError check if error is a duplicate error
func (d MariadbDialect) IsDuplicateError(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "1452"
}

// ConvertValue before load
func (d MariadbDialect) ConvertValue(from push.Value) push.Value {
	return from
}
