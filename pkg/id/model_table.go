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

package id

import (
	"fmt"
	"strings"
)

type table struct {
	name string
}

// NewTable initialize a new Table object
func NewTable(name string) Table {
	return table{name: name}
}

func (t table) Name() string   { return t.name }
func (t table) String() string { return t.name }

type tableList struct {
	len   uint
	slice []Table
	set   set
}

// NewTableList initialize a new TableList object
func NewTableList(tables []Table) TableList {
	set := newSet()
	for _, tab := range tables {
		set.add(tab.Name())
	}
	return tableList{uint(len(tables)), tables, set}
}

func (l tableList) Len() uint                 { return l.len }
func (l tableList) Table(idx uint) Table      { return l.slice[idx] }
func (l tableList) Contains(name string) bool { return l.set.contains(name) }
func (l tableList) String() string {
	switch l.len {
	case 0:
		return ""
	case 1:
		return fmt.Sprint(l.slice[0])
	}
	sb := strings.Builder{}
	fmt.Fprintf(&sb, "%v", l.slice[0])
	for _, tab := range l.slice[1:] {
		fmt.Fprintf(&sb, " %v", tab)
	}
	return sb.String()
}
