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
	// Select Method
	CreateSelect(sel string, where string, limit string, columns string, from string) string
}

func Select(d Dialect, columns []string, distinct bool, schema string, table string, filters map[string]any, where string, limit uint) string {
	// String Builders
	sqlSelect := &strings.Builder{}
	sqlLimit := &strings.Builder{}
	sqlWhere := &strings.Builder{}
	sqlColumns := &strings.Builder{}
	sqlFrom := &strings.Builder{}

	// Build SELECT clause *******************************************
	sqlSelect.WriteString("SELECT")
	if distinct {
		sqlSelect.WriteString(" DISTINCT")
	}
	if len(columns) > 0 {
		for idx, column := range columns {
			if idx > 0 {
				sqlSelect.WriteString(", ")
			}
			sqlSelect.WriteString(" ")
			sqlSelect.WriteString(column)
		}
	} else {
		sqlColumns.WriteString("*")
	}

	// Build FROM clause *********************************************
	sqlFrom.WriteString("FROM ")
	if len(schema) > 0 {
		sqlFrom.WriteString(schema)
		sqlFrom.WriteString(".")
	}
	sqlFrom.WriteString(table)

	// Build LIMIT clause ********************************************
	if limit > 0 {
		sqlLimit.WriteString(d.Limit(limit))
	}

	values := []interface{}{}
	// Build WHERE clause ********************************************
	if len(filters) > 0 || len(where) > 0 {
		sqlWhere.WriteString("WHERE ")
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

	// Assemble the builders in order using the existing method
	sql := d.CreateSelect(sqlSelect.String(), sqlWhere.String(), sqlLimit.String(), sqlColumns.String(), sqlFrom.String())

	if log.Logger.GetLevel() <= zerolog.DebugLevel {
		printSQL := sql
		for i, v := range values {
			printSQL = strings.ReplaceAll(printSQL, d.Placeholder(i+1), fmt.Sprintf("%v", v))
		}
		log.Debug().Msg(fmt.Sprint(printSQL))
	}

	return sql
}
