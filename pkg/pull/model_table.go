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

	"github.com/cgi-fr/jsonline/pkg/jsonline"
	"github.com/rs/zerolog/log"
)

type table struct {
	name    string
	pk      []string
	columns ColumnList
}

type columnList struct {
	len   uint
	slice []Column
}

// NewTable initialize a new Table object
func NewTable(name string, pk []string, columns ColumnList) Table {
	return table{name: name, pk: pk, columns: columns}
}

func (t table) Name() string         { return t.name }
func (t table) PrimaryKey() []string { return t.pk }
func (t table) Columns() ColumnList  { return t.columns }
func (t table) String() string       { return t.name }

func (t table) export(row Row) ExportableRow {
	result := NewExportableRow()
	if t.Columns() == nil || t.Columns().Len() == 0 {
		for k, v := range row {
			result.set(k, v)
		}
		return result
	}
	for i := uint(0); i < t.Columns().Len(); i++ {
		column := t.Columns().Column(i)
		log.Info().Str("column", column.Name()).Str("export", column.Export()).Msg("format")
		key := column.Name()
		val := row[key]

		switch column.Export() {
		case "string":
			result.set(key, jsonline.NewValueString(val))
		case "numeric":
			result.set(key, jsonline.NewValueNumeric(val))
		case "base64":
			result.set(key, jsonline.NewValueBinary(val))
		case "no":
			result.set(key, jsonline.NewValueHidden(val))
		default:
			result.set(key, jsonline.NewValueAuto(val))
		}
	}
	return result
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
	name   string
	export string
}

// NewColumn initialize a new Column object
func NewColumn(name string, export string) Column {
	return column{name, export}
}

func (c column) Name() string   { return c.name }
func (c column) Export() string { return c.export }
