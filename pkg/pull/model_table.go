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
	"sort"

	"github.com/cgi-fr/jsonline/pkg/jsonline"
	"github.com/rs/zerolog/log"
)

func (t *Table) initTemplate() {
	t.template = jsonline.NewTemplate()

	if len(t.Columns) > 0 {
		for _, column := range t.Columns {
			key := column.Name

			switch column.Export {
			case "string":
				t.template.WithString(key)
			case "numeric":
				t.template.WithNumeric(key)
			case "base64":
				t.template.WithBinary(key)
			case "datetime":
				t.template.WithDateTime(key)
			case "timestamp":
				t.template.WithTimestamp(key)
			case "no":
				t.template.WithHidden(key)
			default:
				t.template.WithAuto(key)
			}
		}
	}
}

func (t *Table) export(row Row) ExportedRow {
	if t.template == nil {
		t.initTemplate()
	}

	result := ExportedRow{t.template.CreateRowEmpty()}
	keys := make([]string, 0, len(row))

	if len(t.Columns) > 0 {
		for _, col := range t.Columns {
			keys = append(keys, col.Name)
		}
	} else {
		for k := range row {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys) // this is needed to have a consistent output if no columns is defined by configuration

	for _, k := range keys {
		result.Set(k, row[k])
	}

	keys = keys[:0] // reset slice without unallocating memory

	switch t.ExportMode {
	case ExportModeAll:
		for k := range row {
			if result.GetOrNil(k) == nil {
				keys = append(keys, k)
			}
		}
	case ExportModeOnly: // nothing
	}

	sort.Strings(keys) // this is needed to have a consistent output if no columns is defined by configuration

	for _, k := range keys {
		result.Set(k, row[k])
	}

	return result
}

func (t Table) getKeyValues(row ExportedRow) Row {
	result := Row{}
	for _, key := range t.Keys {
		result[key] = row.GetOrNil(key)
	}

	return result
}

func (t *Table) containsColumn(columnName string) bool {
	for _, col := range t.Columns {
		if col.Name == columnName {
			return true
		}
	}

	return false
}

func (t *Table) addMissingColumns(columnNames ...string) {
	for _, key := range columnNames {
		if !t.containsColumn(key) {
			t.Columns = append(t.Columns, Column{Name: key, Export: "no"})

			log.Warn().
				Str("key", key).
				Interface("table", t.Name).
				Msg("missing required key was automatically added as hidden column")
		}
	}
}
