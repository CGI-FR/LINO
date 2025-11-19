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
	"math"

	"github.com/rs/zerolog/log"
)

type DataSourceInMemory struct {
	tables DataSet
}

func NewDataSourceInMemory(tables DataSet) DataSource {
	return DataSourceInMemory{tables: tables}
}

func (ds DataSourceInMemory) SafeUrl() string { return "mem://test" }
func (ds DataSourceInMemory) Open() error     { return nil }
func (ds DataSourceInMemory) Close() error    { return nil }

func (ds DataSourceInMemory) Read(source Table, filter Filter) (RowSet, error) {
	reader, err := ds.RowReader(source, filter)
	if err != nil {
		return nil, err
	}

	result := RowSet{}
	for reader.Next() {
		result = append(result, reader.Value())
	}

	if reader.Error() != nil {
		return result, fmt.Errorf("%w", reader.Error())
	}

	return result, nil
}

func (ds DataSourceInMemory) RowReader(source Table, filter Filter) (RowReader, error) {
	log.Debug().
		Interface("table", source.Name).
		Interface("filter", filter.Values).
		Interface("select", source.Columns).
		Msg("read from in-memory datasource")

	result := RowSet{}

	allRows, ok := ds.tables[source.Name]
	if !ok {
		return nil, nil
	}
LOOK_FOR_MATCHING_ROWS:
	for _, row := range allRows {
		for key, expected := range filter.Values {
			if row[key] != expected {
				continue LOOK_FOR_MATCHING_ROWS
			}
		}

		if len(source.Columns) == 0 {
			result = append(result, row)
		} else {
			copyr := make(Row, len(source.Columns))
			for _, columns := range source.Columns {
				copyr[columns.Name] = row[columns.Name]
			}

			result = append(result, copyr)
		}

		if filter.Limit > 0 && filter.Limit <= math.MaxInt32 && len(result) >= int(filter.Limit) { //nolint:gosec
			break LOOK_FOR_MATCHING_ROWS
		}
	}

	return &RowReaderInMemory{result}, nil
}

type RowReaderInMemory struct {
	rows RowSet
}

func (rr *RowReaderInMemory) Next() bool { return len(rr.rows) > 0 }
func (rr *RowReaderInMemory) Value() Row {
	row := rr.rows[0]
	rr.rows = rr.rows[1:]

	return row
}
func (rr *RowReaderInMemory) Error() error { return nil }
