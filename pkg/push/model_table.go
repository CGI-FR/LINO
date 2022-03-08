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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cgi-fr/jsonline/pkg/jsonline"
)

type table struct {
	name    string
	pk      []string
	columns ColumnList

	template jsonline.Template
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
	for _, col := range l.slice[1:] {
		fmt.Fprintf(&sb, " -> %v", col)
	}
	return sb.String()
}

type column struct {
	name string
	exp  string
	imp  string
}

// NewColumn initialize a new Column object
func NewColumn(name string, exp string, imp string) Column {
	return column{name, exp, imp}
}

func (c column) Name() string   { return c.name }
func (c column) Export() string { return c.exp }
func (c column) Import() string { return c.imp }

type ImportedRow struct {
	jsonline.Row
}

func (t *table) initTemplate() {
	t.template = jsonline.NewTemplate()

	if t.columns == nil {
		return
	}

	if l := int(t.columns.Len()); l > 0 {
		for idx := 0; idx < l; idx++ {
			col := t.columns.Column(uint(idx))
			key := col.Name()

			switch col.Export() {
			case "string":
				t.template.WithMappedString(key, parseExportType(col.Export()))
			case "numeric":
				t.template.WithMappedNumeric(key, parseExportType(col.Export()))
			case "base64", "binary":
				t.template.WithMappedBinary(key, parseExportType(col.Export()))
			case "datetime":
				t.template.WithMappedDateTime(key, parseExportType(col.Export()))
			case "timestamp":
				t.template.WithMappedTimestamp(key, parseExportType(col.Export()))
			case "no":
				t.template.WithHidden(key)
			default:
				t.template.WithMappedAuto(key, parseExportType(col.Export()))
			}
		}
	}
}

func (t table) Import(row map[string]interface{}) ImportedRow {
	if t.template == nil {
		t.initTemplate()
	}

	result := ImportedRow{t.template.CreateRowEmpty()}
	_ = result.Import(row)
	return result
}

func parseExportType(exp string) jsonline.RawType {
	switch exp {
	case "int":
		return int(0)
	case "int64":
		return int64(0)
	case "int32":
		return int32(0)
	case "int16":
		return int16(0)
	case "int8":
		return int8(0)
	case "uint":
		return uint(0)
	case "uint64":
		return uint64(0)
	case "uint32":
		return uint32(0)
	case "uint16":
		return uint16(0)
	case "uint8":
		return uint8(0)
	case "float64":
		return float64(0)
	case "float32":
		return float32(0)
	case "bool":
		return false
	case "byte":
		return byte(0)
	case "rune":
		return rune(' ')
	case "string":
		return ""
	case "[]byte":
		return []byte{}
	case "time.Time":
		return time.Time{}
	case "json.Number":
		return json.Number("")
	default:
		return nil
	}
}
