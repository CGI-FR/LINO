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

	"github.com/cgi-fr/lino/internal/infra/commonsql"
	"github.com/cgi-fr/lino/pkg/push"
	"github.com/stretchr/testify/assert"
)

func TestAppendColumnToSQLMariaDBWithPreserveBlank(t *testing.T) {
	t.Parallel()
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

	err := appendColumnToSQL(column, sql, MariadbDialect{innerDialect: commonsql.MariadbDialect{}}, 0)
	assert.NotNil(t, err)
}

func TestAppendColumnToSQLMariaDB(t *testing.T) {
	t.Parallel()
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

	d := MariadbDialect{innerDialect: commonsql.MariadbDialect{}}

	err := appendColumnToSQL(column, sql, d, 0)
	assert.Nil(t, err)

	assert.Equal(t, "`column`=?", sql.String())
}
