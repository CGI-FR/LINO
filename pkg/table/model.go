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

package table

// DBInfo holds the type of a column, adding length or size precision if columns fixed thoses info.
type DBInfo struct {
	Type      string
	Length    int64
	Size      int64
	Precision int64
	ByteBased bool
	Preserve  string
}

// Column holds the name of a column.
type Column struct {
	Name   string
	Export string
	Import string
	DBInfo DBInfo
}

type ExportMode byte

const (
	ExportModeOnly ExportMode = iota
	ExportModeAll
)

// Table holds a name (table name) and a list of keys (table columns).
type Table struct {
	Name       string
	Keys       []string
	Columns    []Column
	ExportMode ExportMode
}

// Error is the error type returned by the domain
type Error struct {
	Description string
}

func (e *Error) Error() string {
	return e.Description
}

// TableCount is the number of lines in a table
type TableCount struct {
	Table Table
	Count int
}
