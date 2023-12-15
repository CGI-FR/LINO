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

package rdbms

import "fmt"

// PostgresDialect implement postgres SQL variations
type PostgresDialect struct{}

func (pgd PostgresDialect) Placeholder(position int) string {
	return fmt.Sprintf("$%d", position)
}

func (pgd PostgresDialect) Limit(limit uint) string {
	return fmt.Sprintf("LIMIT %d", limit)
}

// CreateSelect generate a SQL request in the correct order.
func (pgd PostgresDialect) CreateSelect(sel string, where string, limit string, columns string, from string) string {
	return fmt.Sprintf("%s %s %s %s %s", sel, columns, from, where, limit)
}
