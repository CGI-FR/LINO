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

type match struct {
	key   string
	value interface{}
}

func NewFilterExporter(original RowExporter, filters RowReader) (RowExporter, error) {
	result := FilterExporter{original: original, keys: map[string]int{}}
	for filters.Next() {
		result.addFilter(filters.Value())
	}
	return &result, filters.Error()
}

type FilterExporter struct {
	original RowExporter
	keys     map[string]int
	values   []map[interface{}]struct{}
}

func (e *FilterExporter) addFilter(row Row) {
	for k, v := range row {
		index, exist := e.keys[k]
		if !exist {
			index = len(e.values)
			e.keys[k] = index
			e.values = (append(e.values, map[interface{}]struct{}{}))
		}
		e.values[index][v] = struct{}{}
	}
}

func (e *FilterExporter) Export(row ExportedRow) error {
	for key, index := range e.keys {
		v := row.GetOrNil(key)
		_, ok := e.values[index][v]
		if ok {
			err := e.original.Export(row)
			if err != nil {
				return err
			}

			continue
		}
	}
	return nil
}
