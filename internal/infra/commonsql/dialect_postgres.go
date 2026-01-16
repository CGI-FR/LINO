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

// PostgresDialect implement postgres SQL variations
type PostgresDialect struct{}

func (pgd PostgresDialect) Placeholder(position int) string {
	return fmt.Sprintf("$%d", position)
}

func (pgd PostgresDialect) Limit(limit uint) string {
	return fmt.Sprintf("LIMIT %d", limit)
}

// From clause
func (pgd PostgresDialect) From(tableName string, schemaName string) string {
	tableName = pgd.Quote(tableName)
	if strings.TrimSpace(schemaName) == "" {
		return fmt.Sprintf("FROM %s", tableName)
	}
	schemaName = pgd.Quote(schemaName)
	return fmt.Sprintf("FROM %s.%s", schemaName, tableName)
}

// Where clause
func (pgd PostgresDialect) Where(where string) string {
	if strings.TrimSpace(where) == "" {
		return ""
	}

	return fmt.Sprintf("WHERE %s", where)
}

// Select clause
func (pgd PostgresDialect) Select(tableName string, schemaName string, where string, distinct bool, columns ...ColumnExportDefinition) string {
	var query strings.Builder

	query.WriteString("SELECT ")

	if distinct {
		query.WriteString("DISTINCT ")
	}

	if names := Names(columns); len(names) > 0 {
		for i := range columns {
			if columns[i].OnlyPresence {
				names[i] = pgd.selectPresence(names[i])
			} else {
				names[i] = pgd.Quote(names[i])
			}
		}
		query.WriteString(strings.Join(names, ", "))
	} else {
		query.WriteRune('*')
	}

	query.WriteRune(' ')
	query.WriteString(pgd.From(tableName, schemaName))
	query.WriteRune(' ')
	query.WriteString(pgd.Where(where))

	return query.String()
}

// SelectLimit clause
func (pgd PostgresDialect) SelectLimit(tableName string, schemaName string, where string, distinct bool, limit uint, columns ...ColumnExportDefinition) string {
	var query strings.Builder

	query.WriteString("SELECT ")

	if distinct {
		query.WriteString("DISTINCT ")
	}

	if names := Names(columns); len(names) > 0 {
		for i := range names {
			if columns[i].OnlyPresence {
				names[i] = pgd.selectPresence(names[i])
			} else {
				names[i] = pgd.Quote(names[i])
			}
		}
		query.WriteString(strings.Join(names, ", "))
	} else {
		query.WriteRune('*')
	}

	query.WriteRune(' ')
	query.WriteString(pgd.From(tableName, schemaName))
	query.WriteRune(' ')
	query.WriteString(pgd.Where(where))
	query.WriteRune(' ')
	query.WriteString(pgd.Limit(limit))

	return query.String()
}

func (pgd PostgresDialect) Quote(id string) string {
	var sb strings.Builder

	sb.Grow(len(id) + 2)
	sb.WriteRune('"')
	sb.WriteString(strings.TrimSpace(id))
	sb.WriteRune('"')

	return sb.String()
}

// CreateSelect generate a SQL request in the correct order.
func (pgd PostgresDialect) CreateSelect(sel string, where string, limit string, columns string, from string) string {
	return fmt.Sprintf("%s %s %s %s %s", sel, columns, from, where, limit)
}

func (pgd PostgresDialect) selectPresence(column string) string {
	return fmt.Sprintf("CASE WHEN (%s IS NOT NULL) THEN TRUE ELSE NULL END AS %s", pgd.Quote(column), pgd.Quote(column))
}

// BlankTest implements SQLDialect.
func (pgd PostgresDialect) BlankTest(column string) string {
	return fmt.Sprintf("TRIM(%s) = ''", pgd.Quote(column))
}

// EmptyTest implements SQLDialect.
func (pgd PostgresDialect) EmptyTest(column string) string {
	return fmt.Sprintf("%s = ''", pgd.Quote(column))
}

// EnableConstraintsStatement generate statments to activate constraintes
func (pgd PostgresDialect) EnableConstraintsStatement(tableName string) string {
	return fmt.Sprintf("ALTER TABLE %s ENABLE TRIGGER ALL", pgd.Quote(tableName))
}

// DisableConstraintsStatement generate statments to deactivate constraintes
func (pgd PostgresDialect) DisableConstraintsStatement(tableName string) string {
	return fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER ALL", pgd.Quote(tableName))
}

// TruncateStatement generate statement to truncat table content
func (pgd PostgresDialect) TruncateStatement(tableName string) string {
	return fmt.Sprintf("TRUNCATE TABLE %s CASCADE", pgd.Quote(tableName))
}
