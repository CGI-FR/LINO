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
	"strings"
	"testing"

	"github.com/cgi-fr/lino/pkg/push"
	"github.com/stretchr/testify/assert"
)

func TestAppendColumnToSQLPGPreserveNothing(t *testing.T) {
	sql := &strings.Builder{}
	column := ValueDescriptor{
		name: "column",
		column: push.NewColumn(
			"column",
			"",
			"",
			0,
			false,
			false,

			push.PreserveNothing,
		),
	}

	appendColumnToSQLPG(sql, column, PostgresDialect{}, 0, 2)
	assert.Equal(t, "column=$1, ", sql.String())

}

func TestAppendColumnToSQLPGPreserveBlank(t *testing.T) {
	sql := &strings.Builder{}
	column := ValueDescriptor{
		name: "column",
		column: push.NewColumn(
			"column",
			"",
			"",
			0,
			false,
			false,

			push.PreserveBlank,
		),
	}

	appendColumnToSQLPG(sql, column, PostgresDialect{}, 0, 2)

	expected := "column = CASE WHEN (column IS NULL) OR (TRIM(column) = '') THEN column ELSE $1 END, "

	assert.Equal(t, expected, sql.String())

}
