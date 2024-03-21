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
func (e *MariadbDataDestinationFactory) New(url string, schema string, options ...push.DataDestinationOption) push.DataDestination {
	return NewSQLDataDestination(url, schema, MariadbDialect{}, options...)
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
func (d MariadbDialect) InsertStatement(tableName string, selectValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor) {
	protectedColumns := []string{}
	for _, value := range selectValues {
		protectedColumns = append(protectedColumns, fmt.Sprintf("`%s`", value.name))
	}

	sql := &strings.Builder{}
	if len(primaryKeys) > 0 {
		sql.WriteString("INSERT IGNORE INTO ")
	} else {
		sql.WriteString("INSERT INTO ")
	}
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
	sql.WriteString(")")

	return sql.String(), selectValues
}

func (d MariadbDialect) UpdateStatement(tableName string, selectValues []ValueDescriptor, whereValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor, err *push.Error) {
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
		sql.Write([]byte(" WHERE "))
	} else {
		return "", nil, &push.Error{Description: fmt.Sprintf("can't update table [%s] because no primary key is defined", tableName)}
	}
	for index, pk := range whereValues {
		headers = append(headers, pk)

		sql.WriteString(pk.name)
		sql.WriteString("=")
		sql.WriteString(d.Placeholder(len(selectValues) + index + 1))
		if index+1 < len(whereValues) {
			sql.WriteString(" AND ")
		}
	}

	return sql.String(), headers, nil
}

// IsDuplicateError check if error is a duplicate error
func (d MariadbDialect) IsDuplicateError(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "1452"
}

// ConvertValue before load
func (d MariadbDialect) ConvertValue(from push.Value, descriptor ValueDescriptor) push.Value {
	return from
}

func (d MariadbDialect) CanDisableIndividualConstraints() bool {
	return false
}

func (d MariadbDialect) ReadConstraintsStatement(tableName string) string {
	panic(fmt.Errorf("Not implemented"))
}

func (d MariadbDialect) DisableConstraintStatement(tableName string, constraintName string) string {
	panic(fmt.Errorf("Not implemented"))
}

func (d MariadbDialect) EnableConstraintStatement(tableName string, constraintName string) string {
	panic(fmt.Errorf("Not implemented"))
}
