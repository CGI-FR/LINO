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

	_ "github.com/microsoft/go-mssqldb"
	mssql "github.com/microsoft/go-mssqldb"
)

// SQLServerDataDestinationFactory exposes methods to create new SQLServer pullers.
type SQLServerDataDestinationFactory struct{}

// NewSQLServerDataDestinationFactory creates a new SQLServer datadestination factory.
func NewSQLServerDataDestinationFactory() *SQLServerDataDestinationFactory {
	return &SQLServerDataDestinationFactory{}
}

// New return a SQLServer pusher
func (e *SQLServerDataDestinationFactory) New(url string, schema string) push.DataDestination {
	return NewSQLDataDestination(url, schema, SQLServerDialect{})
}

// SQLServerDialect inject SQLServer variations
type SQLServerDialect struct{}

// Placeholde return the variable format for SQLServer
func (d SQLServerDialect) Placeholder(position int) string {
	return fmt.Sprintf("@p%d", position)
}

// EnableConstraintsStatement generate statments to activate constraintes
func (d SQLServerDialect) EnableConstraintsStatement(tableName string) string {
	return fmt.Sprintf("ALTER TABLE %s CHECK CONSTRAINT ALL", tableName)
}

// DisableConstraintsStatement generate statments to deactivate constraintes
func (d SQLServerDialect) DisableConstraintsStatement(tableName string) string {
	return fmt.Sprintf("ALTER TABLE %s NOCHECK CONSTRAINT ALL", tableName)
}

// TruncateStatement generate statement to truncate table content for SQL Server
func (d SQLServerDialect) TruncateStatement(tableName string) string {
	return fmt.Sprintf("DELETE FROM %s", tableName)
}

// InsertStatement generates an insert statement for SQL Server
func (d SQLServerDialect) InsertStatement(tableName string, selectValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor) {
	protectedColumns := []string{}
	for _, value := range selectValues {
		protectedColumns = append(protectedColumns, fmt.Sprintf("[%s]", value.name))
	}

	sql := &strings.Builder{}
	sql.WriteString("INSERT INTO ")
	sql.WriteString(tableName)
	sql.WriteString("(")
	sql.WriteString(strings.Join(protectedColumns, ","))
	sql.WriteString(") VALUES (")
	for i := 1; i <= len(selectValues); i++ {
		sql.WriteString(d.Placeholder(i)) // Assuming Placeholder is a method that returns the appropriate placeholder for SQL Server, like "?"
		if i < len(selectValues) {
			sql.WriteString(", ")
		}
	}
	sql.WriteString(")")

	return sql.String(), selectValues
}

func (d SQLServerDialect) UpdateStatement(tableName string, selectValues []ValueDescriptor, whereValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor, err *push.Error) {
	sql := &strings.Builder{}
	sql.WriteString("UPDATE ")
	sql.WriteString(tableName)
	sql.WriteString(" SET ")

	for index, column := range selectValues {
		if isAPrimaryKey(column.name, primaryKeys) {
			isInWhere := false
			for _, pk := range whereValues {
				if column.name == pk.name {
					isInWhere = true
					break
				}
			}
			if !isInWhere {
				continue
			}
		}

		headers = append(headers, column)

		errColumn := appendColumnToSQL(column, sql, d, index)
		if errColumn != nil {
			return "", nil, errColumn
		}

		if index+1 < len(selectValues) {
			sql.WriteString(", ")
		}
	}

	if len(headers) == 0 {
		return "", nil, &push.Error{Description: fmt.Sprintf("can't update table [%s] because no columns are selected", tableName)}
	}

	if len(whereValues) > 0 {
		sql.WriteString(" WHERE ")
	} else {
		return "", nil, &push.Error{Description: fmt.Sprintf("can't update table [%s] because no primary key is defined", tableName)}
	}

	for index, pk := range whereValues {
		headers = append(headers, pk)

		sql.WriteString(pk.name)
		sql.WriteString("=")
		sql.WriteString(d.Placeholder(len(selectValues) + index + 1))
		if index+1 < len(whereValues) {
			sql.Write([]byte(" AND "))
		}
	}

	return sql.String(), headers, nil
}

// IsDuplicateError check if error is a duplicate error
func (d SQLServerDialect) IsDuplicateError(err error) bool {
	msErr, ok := err.(mssql.Error)
	return ok && msErr.Number == 2627 // Check violation number in https://github.com/microsoft/go-mssqldb/blob/main/error.go
}

// ConvertValue before load
func (d SQLServerDialect) ConvertValue(from push.Value, descriptor ValueDescriptor) push.Value {
	return from
}

func (d SQLServerDialect) CanDisableIndividualConstraints() bool {
	return false
}

func (d SQLServerDialect) ReadConstraintsStatement(tableName string) string {
	panic(fmt.Errorf("Not implemented"))
}

func (d SQLServerDialect) DisableConstraintStatement(tableName string, constraintName string) string {
	panic(fmt.Errorf("Not implemented"))
}

func (d SQLServerDialect) EnableConstraintStatement(tableName string, constraintName string) string {
	panic(fmt.Errorf("Not implemented"))
}

func (d SQLServerDialect) SupportPreserve() []string {
	return []string{
		string(push.PreserveNothing),
	}
}

// BlankTest implements SQLDialect.
func (d SQLServerDialect) BlankTest(name string) string {
	panic("unimplemented")
}

func (d SQLServerDialect) EmptyTest(name string) string {
	panic("unimplemented")
}
