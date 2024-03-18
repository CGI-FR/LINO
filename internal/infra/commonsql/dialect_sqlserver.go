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

// SQLServerDialect implement SQLServer SQL variations
type SQLServerDialect struct{} //nolint:golint,revive

func (sd SQLServerDialect) Placeholder(position int) string {
	return fmt.Sprintf("@p%d", position)
}

// Limit method is adjusted to be compatible with SQL Server
func (sd SQLServerDialect) Limit(limit uint) string {
	return fmt.Sprintf("TOP %d", limit)
}

// From clause
func (sd SQLServerDialect) From(tableName string, schemaName string) string {
	if strings.TrimSpace(schemaName) == "" {
		return fmt.Sprintf("FROM %s", tableName)
	}

	return fmt.Sprintf("FROM %s.%s", schemaName, tableName)
}

// Where clause
func (sd SQLServerDialect) Where(where string) string {
	if strings.TrimSpace(where) == "" {
		return ""
	}

	return fmt.Sprintf("WHERE %s", where)
}

// Select clause
func (sd SQLServerDialect) Select(tableName string, schemaName string, where string, distinct bool, columns ...string) string {
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
	query.WriteString(sd.From(tableName, schemaName))
	query.WriteRune(' ')
	query.WriteString(sd.Where(where))

	return query.String()
}

// SelectLimit clause
func (sd SQLServerDialect) SelectLimit(tableName string, schemaName string, where string, distinct bool, limit uint, columns ...string) string {
	var query strings.Builder

	query.WriteString("SELECT ")

	if distinct {
		query.WriteString("DISTINCT ")
	}

	query.WriteString(sd.Limit(limit))
	query.WriteRune(' ')

	if len(columns) > 0 {
		query.Write([]byte(" "))
		query.WriteString(strings.Join(columns, ", "))
	} else {
		query.WriteRune('*')
	}

	query.WriteRune(' ')
	query.WriteString(sd.From(tableName, schemaName))
	query.WriteRune(' ')
	query.WriteString(sd.Where(where))

	return query.String()
}

func (sd SQLServerDialect) Quote(id string) string {
	return id
}

// CreateSelect generate a SQL request in the correct order
func (sd SQLServerDialect) CreateSelect(sel string, where string, limit string, columns string, from string) string {
	return fmt.Sprintf("%s %s %s %s %s", sel, limit, columns, from, where)
}
