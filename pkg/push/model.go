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

// Table from which to push data.
type Table interface {
	Name() string
	PrimaryKey() []string
}

// Plan describe how to push data
type Plan interface {
	FirstTable() Table
	RelationsFromTable(table Table) map[string]Relation
	Tables() []Table
}

// Relation between two tables.
type Relation interface {
	Name() string
	Parent() Table
	Child() Table
	OppositeOf(table Table) Table
}

// Value is an untyped data.
type Value interface{}

// Row of data.
type Row map[string]Value

// Error is the error type returned by the domain
type Error struct {
	Description string
}

func (e *Error) Error() string {
	return e.Description
}

// StopIteratorError signal the end of iterator
type StopIteratorError struct{}
