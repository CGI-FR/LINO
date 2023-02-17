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
type PostgresDataDestinationFactory struct{}

// NewPostgresDataDestinationFactory creates a new postgres datadestination factory.
func NewPostgresDataDestinationFactory() *PostgresDataDestinationFactory {
	return &PostgresDataDestinationFactory{}
}

// New return a Postgres pusher
func (e *PostgresDataDestinationFactory) New(url string, schema string) push.DataDestination {
	return NewSQLDataDestination(url, schema, PostgresDialect{})
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
func (d PostgresDialect) InsertStatement(tableName string, selectValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor) {
	protectedColumns := []string{}
	for _, c := range selectValues {
		protectedColumns = append(protectedColumns, fmt.Sprintf("\"%s\"", c.name))
	}

	sql := &strings.Builder{}
	sql.WriteString("INSERT INTO ")
	sql.WriteString(tableName)
	sql.WriteString("(")
	sql.WriteString(strings.Join(protectedColumns, ","))
	sql.WriteString(") VALUES (")
	for i := 1; i <= len(selectValues); i++ {
		sql.WriteString(d.Placeholder(i))
		if i < len(selectValues) {
			sql.WriteString(", ")
		}
	}
	if len(primaryKeys) > 0 {
		sql.WriteString(") ON CONFLICT (")
		sql.WriteString(strings.Join(primaryKeys, ","))
		sql.WriteString(") DO NOTHING")
	} else {
		sql.WriteString(")")
	}

	return sql.String(), selectValues
}

func (d PostgresDialect) UpdateStatement(tableName string, selectValues []ValueDescriptor, whereValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor, err *push.Error) {
	sql := &strings.Builder{}
	sql.WriteString("UPDATE ")
	sql.WriteString(tableName)
	sql.WriteString(" SET ")

	for index, column := range selectValues {
		// don't update primary key, except if it's in whereValues
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

		sql.WriteString(column.name)
		sql.WriteString("=")
		sql.WriteString(d.Placeholder(index + 1))
		if index+1 < len(selectValues) {
			sql.WriteString(", ")
		}
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
func (d PostgresDialect) IsDuplicateError(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "23505"
}

// ConvertValue before load
func (d PostgresDialect) ConvertValue(from push.Value) push.Value {
	return from
}

func (d PostgresDialect) CanDisableIndividualConstraints() bool {
	return false
}

func (d PostgresDialect) ReadConstraintsStatement(tableName string) string {
	panic(fmt.Errorf("Not implemented"))
}

func (d PostgresDialect) DisableContraintStatement(tableName string, constraintName string) string {
	panic(fmt.Errorf("Not implemented"))
}

func (d PostgresDialect) EnableConstraintStatement(tableName string, constraintName string) string {
	panic(fmt.Errorf("Not implemented"))
}
