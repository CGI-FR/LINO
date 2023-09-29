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

package analyse

// Do performe statistics on datasource.
func Do(ds DataSource, ex Extractor, analyser Analyser) error {
	iterator := NewColumnIterator(ds, ex)
	return analyser.Analyse(iterator)
}

type ColumnIterator struct {
	tables []string
	column []string
	DataSource
	Extractor
}

func NewColumnIterator(ds DataSource, ex Extractor) *ColumnIterator {
	return &ColumnIterator{
		tables:     []string{},
		column:     []string{},
		DataSource: ds,
		Extractor:  ex,
	}
}

func (ci *ColumnIterator) BaseName() string { return ci.Name() }

// Next return true if there is more column to iterate over.
func (ci *ColumnIterator) Next() bool {
	if len(ci.tables) == 0 {
		ci.tables = ci.ListTables()
		if len(ci.tables) == 0 {
			return false
		}
		ci.column = ci.ListColumn(ci.tables[0])
		if len(ci.column) > 0 {
			return true
		}
	}

	if len(ci.column) > 1 {
		ci.column = ci.column[1:]
		return true
	}

	for len(ci.tables) > 1 {
		ci.tables = ci.tables[1:]
		ci.column = ci.DataSource.ListColumn(ci.tables[0])
		if len(ci.column) > 0 {
			return true
		}
	}
	return false
}

// Value return the column content.
func (ci *ColumnIterator) Value() ([]interface{}, string, string, error) {
	return ci.ExtractValues(ci.tables[0], ci.column[0]), ci.tables[0], ci.column[0], nil
}
