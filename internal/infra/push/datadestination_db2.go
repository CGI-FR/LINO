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

//go:build db2
// +build db2

package push

import (
	"fmt"
	"strings"
	"time"

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
func (d Db2Dialect) InsertStatement(tableName string, selectValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor) {
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
	sql.WriteString(")")

	return sql.String(), selectValues
}

// UpdateStatement
func (d Db2Dialect) UpdateStatement(tableName string, selectValues []ValueDescriptor, whereValues []ValueDescriptor, primaryKeys []string) (statement string, headers []ValueDescriptor, err *push.Error) {
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

		errColumn := appendColumnToSQL(column, sql, d, index)
		if errColumn != nil {
			return "", nil, errColumn
		}

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
func (d Db2Dialect) IsDuplicateError(err error) bool {
	// -803
	return strings.Contains(err.Error(), "-803")
}

// ConvertValue before load
func (d Db2Dialect) ConvertValue(from push.Value, descriptor ValueDescriptor) push.Value {
	// FIXME: Workaround to parse time from json
	aTime, err := time.Parse("2006-01-02T15:04:05.999Z07:00", fmt.Sprintf("%v", from))
	if err != nil {
		return from
	} else {
		return aTime
	}
}

func (d Db2Dialect) CanDisableIndividualConstraints() bool {
	return false
}

func (d Db2Dialect) ReadConstraintsStatement(tableName string) string {
	panic(fmt.Errorf("Not implemented"))
}

func (d Db2Dialect) DisableConstraintStatement(tableName string, constraintName string) string {
	panic(fmt.Errorf("Not implemented"))
}

func (d Db2Dialect) EnableConstraintStatement(tableName string, constraintName string) string {
	panic(fmt.Errorf("Not implemented"))
}

func (d Db2Dialect) SupportPreserve() bool {
	return false
}
