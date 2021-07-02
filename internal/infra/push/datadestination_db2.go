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

// +build db2

package push

import (
	"fmt"
	"strings"

	// import db2 connector
	_ "github.com/ibmdb/go_ibm_db"

	"github.com/cgi-fr/lino/pkg/push"
)

// Db2DataDestinationFactory exposes methods to create new Db2 extractors.
type Db2DataDestinationFactory struct{}

// NewDb2DataDestinationFactory creates a new Db2 datadestination factory.
func NewDb2DataDestinationFactory() *Db2DataDestinationFactory {
	return &Db2DataDestinationFactory{}
}

// New return a Db2 pusher
func (e *Db2DataDestinationFactory) New(url string, schema string) push.DataDestination {
	return NewSQLDataDestination(url, schema, Db2Dialect{})
}

// Db2Dialect inject oracle variations
type Db2Dialect struct{}

// Placeholde return the variable format for postgres
func (d Db2Dialect) Placeholder(position int) string {
	return "?"
}

// EnableConstraintsStatement generate statments to activate constraintes
func (d Db2Dialect) EnableConstraintsStatement(tableName string) string {
	panic(fmt.Errorf("Not implemented"))
}

// DisableConstraintsStatement generate statments to deactivate constraintes
func (d Db2Dialect) DisableConstraintsStatement(tableName string) string {
	panic(fmt.Errorf("Not implemented"))
}

// TruncateStatement generate statement to truncat table content
func (d Db2Dialect) TruncateStatement(tableName string) string {
	return fmt.Sprintf("TRUNCATE TABLE %s IMMEDIATE", tableName)
}

// InsertStatement generate insert statement
func (d Db2Dialect) InsertStatement(tableName string, columns []string, values []string, primaryKeys []string) string {
	protectedColumns := []string{}
	for _, c := range columns {
		protectedColumns = append(protectedColumns, fmt.Sprintf("\"%s\"", c))
	}
	return fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", tableName, strings.Join(protectedColumns, ","), strings.Join(values, ","))
}

// UpdateStatement
func (d Db2Dialect) UpdateStatement(tableName string, columns []string, uValues []string, primaryKeys []string, pValues []string) (string, *push.Error) {
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
func (d Db2Dialect) IsDuplicateError(err error) bool {
	// -803
	return strings.Contains(err.Error(), "-803")
}

// ConvertValue before load
func (d Db2Dialect) ConvertValue(from push.Value) push.Value {
	return from
}
