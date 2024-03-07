// Copyright (C) 2023 CGI France
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

package commonsql

import (
	"fmt"
	"strings"
)

// OracleDialect implement Oracle SQL variations
type OracleDialect struct{}

func (od OracleDialect) Placeholder(position int) string {
	return fmt.Sprintf(":v%d", position)
}

func (od OracleDialect) Limit(limit uint) string {
	return fmt.Sprintf(" AND rownum <= %d", limit)
}

// From clause
func (od OracleDialect) From(tableName string, schemaName string) string {
	tableName = od.Quote(tableName)
	if strings.TrimSpace(schemaName) == "" {
		return fmt.Sprintf("FROM %s", tableName)
	}
	schemaName = od.Quote(schemaName)
	return fmt.Sprintf("FROM %s.%s", schemaName, tableName)
}

// Where clause
func (od OracleDialect) Where(where string) string {
	if strings.TrimSpace(where) == "" {
		return "WHERE 1=1"
	}

	return fmt.Sprintf("WHERE %s", where)
}

// Select clause
func (od OracleDialect) Select(tableName string, schemaName string, where string, distinct bool, columns ...string) string {
	var query strings.Builder

	query.WriteString("SELECT ")

	if distinct {
		query.WriteString("DISTINCT ")
	}

	if len(columns) > 0 {
		query.WriteString(strings.Join(columns, ", "))
	} else {
		query.WriteRune('*')
	}

	query.WriteRune(' ')
	query.WriteString(od.From(tableName, schemaName))
	query.WriteRune(' ')
	query.WriteString(od.Where(where))

	return query.String()
}

// SelectLimit clause
func (od OracleDialect) SelectLimit(tableName string, schemaName string, where string, distinct bool, limit uint, columns ...string) string {
	var query strings.Builder

	query.WriteString("SELECT ")

	if distinct {
		query.WriteString("DISTINCT ")
	}

	if len(columns) > 0 {
		query.WriteString(strings.Join(columns, ", "))
	} else {
		query.WriteRune('*')
	}

	query.WriteRune(' ')
	query.WriteString(od.From(tableName, schemaName))
	query.WriteRune(' ')
	query.WriteString(od.Where(where))
	query.WriteRune(' ')
	query.WriteString(od.Limit(limit))

	return query.String()
}

func (od OracleDialect) Quote(id string) string {
	var sb strings.Builder

	sb.Grow(len(id) + 2)
	sb.WriteRune('"')
	sb.WriteString(id)
	sb.WriteRune('"')

	return sb.String()
}

// CreateSelect generate a SQL request in the correct order.
func (od OracleDialect) CreateSelect(sel string, where string, limit string, columns string, from string) string {
	return fmt.Sprintf("%s %s %s %s %s", sel, columns, from, where, limit)
}
