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
func (od OracleDialect) Select(tableName string, schemaName string, where string, distinct bool, columns ...ColumnExportDefinition) string {
	var query strings.Builder

	query.WriteString("SELECT ")

	if distinct {
		query.WriteString("DISTINCT ")
	}

	if names := Names(columns); len(names) > 0 {
		for i := range names {
			if columns[i].OnlyPresence {
				names[i] = od.selectPresence(names[i])
			} else {
				names[i] = od.Quote(names[i])
			}
		}
		query.WriteString(strings.Join(names, ", "))
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
func (od OracleDialect) SelectLimit(tableName string, schemaName string, where string, distinct bool, limit uint, columns ...ColumnExportDefinition) string {
	var query strings.Builder

	query.WriteString("SELECT ")

	if distinct {
		query.WriteString("DISTINCT ")
	}

	if names := Names(columns); len(names) > 0 {
		for i := range names {
			if columns[i].OnlyPresence {
				names[i] = od.selectPresence(names[i])
			} else {
				names[i] = od.Quote(names[i])
			}
		}
		query.WriteString(strings.Join(names, ", "))
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

func (od OracleDialect) selectPresence(column string) string {
	return fmt.Sprintf("CASE WHEN %s IS NOT NULL THEN 'TRUE' ELSE NULL END AS %s", od.Quote(column), od.Quote(column))
}

// BlankTest generate a SQL test to check if a column is blank (spaces only)
func (od OracleDialect) BlankTest(column string) string {
	return fmt.Sprintf("TRIM(%s) IS NULL", od.Quote(column))
}

// EmptyTest generate a SQL test to check if a column is empty (zero length)
func (od OracleDialect) EmptyTest(column string) string {
	return fmt.Sprintf("%s IS NULL", od.Quote(column))
}

// EnableConstraintsStatement generate statments to activate constraintes
func (od OracleDialect) EnableConstraintsStatement(tableName string) string {
	schemaAndTable := strings.Split(tableName, ".")
	sql := &strings.Builder{}
	sql.WriteString(
		`BEGIN
		 FOR c IN(
		 SELECT c.owner, c.table_name, c.constraint_name
		 FROM user_constraints c
		 CONNECT BY PRIOR c.constraint_name = c.r_constraint_name
		 START WITH c.constraint_name IN (
			SELECT c.constraint_name
			FROM user_constraints c
		 	WHERE c.status = 'DISABLED' AND c.table_name = '`)
	if len(schemaAndTable) == 2 {
		sql.WriteString(schemaAndTable[1])
		sql.WriteString("' AND c.owner = '")
		sql.WriteString(schemaAndTable[0])
		sql.WriteString("'")
	} else {
		sql.WriteString(schemaAndTable[0])
		sql.WriteString("' AND c.owner = sys_context( 'userenv', 'current_schema' )")
	}
	sql.WriteString(`)
		LOOP
			dbms_utility.exec_ddl_statement('alter table "' || c.owner || '"."' || c.table_name || '" disable constraint ' || c.constraint_name);
		END LOOP;
	END;`)
	return sql.String()
}

// DisableConstraintsStatement generate statments to deactivate constraintes
func (od OracleDialect) DisableConstraintsStatement(tableName string) string {
	schemaAndTable := strings.Split(tableName, ".")
	sql := &strings.Builder{}
	sql.WriteString(
		`BEGIN
		 FOR c IN(
		 SELECT c.owner, c.table_name, c.constraint_name
		 FROM user_constraints c
		 CONNECT BY PRIOR c.constraint_name = c.r_constraint_name
		 START WITH c.constraint_name IN (
			SELECT c.constraint_name
			FROM user_constraints c
		 	WHERE c.status = 'ENABLED' AND c.table_name = '`)
	if len(schemaAndTable) == 2 {
		sql.WriteString(schemaAndTable[1])
		sql.WriteString("' AND c.owner = '")
		sql.WriteString(schemaAndTable[0])
		sql.WriteString("'")
	} else {
		sql.WriteString(schemaAndTable[0])
		sql.WriteString("' AND c.owner = sys_context( 'userenv', 'current_schema' )")
	}
	sql.WriteString(`)
		LOOP
			dbms_utility.exec_ddl_statement('alter table "' || c.owner || '"."' || c.table_name || '" disable constraint ' || c.constraint_name);
		END LOOP;
	END;`)
	return sql.String()
}

// TruncateStatement generate statement to truncat table content
func (od OracleDialect) TruncateStatement(tableName string) string {
	schemaAndTable := strings.Split(tableName, ".")
	if len(schemaAndTable) == 2 {
		return fmt.Sprintf("TRUNCATE TABLE %s.%s", od.Quote(schemaAndTable[0]), od.Quote(schemaAndTable[1]))
	}
	return fmt.Sprintf("TRUNCATE TABLE %s", od.Quote(tableName))
}
