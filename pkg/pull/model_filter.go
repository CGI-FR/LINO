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

package pull

import (
	"fmt"
	"strings"
)

type filter struct {
	limit    uint
	values   Row
	where    string
	distinct bool
}

// NewFilter initialize a new Filter object
func NewFilter(limit uint, values Row, where string, distinct bool) Filter {
	return filter{
		limit:    limit,
		values:   values,
		where:    strings.TrimSpace(where),
		distinct: distinct,
	}
}

func (f filter) Limit() uint    { return f.limit }
func (f filter) Values() Row    { return f.values }
func (f filter) Where() string  { return f.where }
func (f filter) Distinct() bool { return f.distinct }

func (f filter) String() string {
	builder := &strings.Builder{}
	cnt := len(f.Values())
	for key, value := range f.Values() {
		fmt.Fprintf(builder, "%v=%v", key, value)
		cnt--
		if cnt > 0 {
			fmt.Fprint(builder, " ")
		}
	}
	if len(f.Values()) == 0 && f.Limit() == 0 {
		fmt.Fprintf(builder, "true")
	}
	if f.Limit() > 0 {
		if len(f.Values()) > 0 {
			fmt.Fprint(builder, " ")
		}
		fmt.Fprintf(builder, "limit %v", f.Limit())
	}
	return builder.String()
}
