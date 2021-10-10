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
	"fmt"
	"strings"
)

type table struct {
	name    string
	pk      []string
	columns ColumnList
}

// NewTable initialize a new Table object
func NewTable(name string, pk []string, columns ColumnList) Table {
	return table{name: name, pk: pk, columns: columns}
}

func (t table) Name() string         { return t.name }
func (t table) PrimaryKey() []string { return t.pk }
func (t table) Columns() ColumnList  { return t.columns }
func (t table) String() string       { return t.name }

type columnList struct {
	len   uint
	slice []Column
}

// NewColumnList initialize a new ColumnList object
func NewColumnList(columns []Column) ColumnList {
	return columnList{uint(len(columns)), columns}
}

func (l columnList) Len() uint              { return l.len }
func (l columnList) Column(idx uint) Column { return l.slice[idx] }
func (l columnList) String() string {
	switch l.len {
	case 0:
		return ""
	case 1:
		return fmt.Sprint(l.slice[0])
	}
	sb := strings.Builder{}
	fmt.Fprintf(&sb, "%v", l.slice[0])
	for _, rel := range l.slice[1:] {
		fmt.Fprintf(&sb, " -> %v", rel)
	}
	return sb.String()
}

type column struct {
	name string
	imp  string
}

// NewColumn initialize a new Column object
func NewColumn(name string, imp string) Column {
	return column{name, imp}
}

func (c column) Name() string   { return c.name }
func (c column) Import() string { return c.imp }
