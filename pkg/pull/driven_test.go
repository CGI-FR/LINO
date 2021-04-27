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

package pull_test

import (
	"fmt"

	"github.com/cgi-fr/lino/pkg/pull"
)

// MemoryDataIterator
type MemoryDataIterator struct {
	rows    []pull.Row
	current pull.Row
}

func (di *MemoryDataIterator) Next() bool {
	if len(di.rows) > 0 {
		di.current = di.rows[0]
		di.rows = di.rows[1:]
		return true
	}

	return false
}

func (di *MemoryDataIterator) Value() pull.Row {
	return di.current
}

func (di *MemoryDataIterator) Error() *pull.Error {
	return nil
}

// MemoryDataSource mocks DataSource.
type MemoryDataSource struct {
	data map[string][]pull.Row
}

func copyRow(r pull.Row) pull.Row {
	copy := pull.Row{}
	for key, value := range r {
		copy[key] = value
	}
	return copy
}

func (ds *MemoryDataSource) Open() *pull.Error {
	return nil
}

func (ds *MemoryDataSource) Close() *pull.Error {
	return nil
}

func (ds *MemoryDataSource) RowReader(source pull.Table, filter pull.Filter) (pull.RowReader, *pull.Error) {
	rows, ok := ds.data[source.Name()]
	result := []pull.Row{}
	if ok {
	loop:
		for _, row := range rows {
			if len(filter.Values()) == 0 {
				result = append(result, copyRow(row))
				if uint(len(result)) == filter.Limit() {
					break
				}
				continue
			}
			for key, value := range filter.Values() {
				if row[key] == value {
					result = append(result, copyRow(row))
					if uint(len(result)) == filter.Limit() {
						break loop
					}
					continue loop
				}
			}
		}
	}
	fmt.Println("SELECT FROM", source.Name(), "WHERE", filter, "\n  returned:", result)
	return &MemoryDataIterator{result, nil}, nil
}

// MemoryRowExporter mock RowExporter.
type MemoryRowExporter struct {
	rows []pull.Row
}

func (re *MemoryRowExporter) Export(r pull.Row) *pull.Error {
	re.rows = append(re.rows, r)
	fmt.Println("Exported:", r)
	return nil
}

// Logger implementation.
type Logger struct{}

// Trace event.
func (l Logger) Trace(msg string) {
	fmt.Printf("[trace] %v\n", msg)
}

// Debug event.
func (l Logger) Debug(msg string) {
	fmt.Printf("[debug] %v\n", msg)
}

// Info event.
func (l Logger) Info(msg string) {
	fmt.Printf("[info]  %v\n", msg)
}

// Warn event.
func (l Logger) Warn(msg string) {
	fmt.Printf("[warn]  %v\n", msg)
}

// Error event.
func (l Logger) Error(msg string) {
	fmt.Printf("[error] %v\n", msg)
}
