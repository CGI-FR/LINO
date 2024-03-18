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

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Dialect interface {
	// Placeholder format variable in query
	Placeholder(int) string
	// Limit format limitation clause
	Limit(uint) string
	// From clause
	From(tableName string, schemaName string) string
	// Where clause
	Where(string) string
	// Select clause
	Select(tableName string, schemaName string, where string, distinct bool, columns ...string) string
	// SelectLimit clause
	SelectLimit(tableName string, schemaName string, where string, distinct bool, limit uint, columns ...string) string
	// Quote identifier
	Quote(id string) string

	// Deprecated
	CreateSelect(sel string, where string, limit string, columns string, from string) string
}

// Build WHERE clause with where key and value to a string
func GetWhereSQLAndValues(filters map[string]any, where string, d Dialect) (string, []interface{}) {
	values := []interface{}{}
	sqlWhere := &strings.Builder{}

	if len(filters) > 0 || len(where) > 0 {
		whereContentFlag := false
		for key, value := range filters {
			sqlWhere.WriteString(key)
			values = append(values, value)
			sqlWhere.WriteString("=")
			sqlWhere.WriteString(d.Placeholder(len(values)))
			if len(values) < len(filters) {
				sqlWhere.WriteString(" AND ")
			}
			whereContentFlag = true
		}

		if strings.TrimSpace(where) != "" {
			if whereContentFlag {
				sqlWhere.WriteString(" AND ")
			}
			sqlWhere.WriteString(where)
			whereContentFlag = true
		}

		if !whereContentFlag {
			sqlWhere.WriteString(" 1=1 ")
		}
	}
	return sqlWhere.String(), values
}

// When log level is equal or more than debug level, function will log all the SQL Query
func LogSQLQuery(sql string, values []interface{}, d Dialect) {
	if log.Logger.GetLevel() <= zerolog.DebugLevel {
		printSQL := sql
		for i, v := range values {
			printSQL = strings.ReplaceAll(printSQL, d.Placeholder(i+1), fmt.Sprintf("%v", v))
		}
		log.Debug().Msg(fmt.Sprint(printSQL))
	}
}
