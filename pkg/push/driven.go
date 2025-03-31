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

// DataDestinationFactory exposes methods to create new datadestinations.
type DataDestinationFactory interface {
	New(url string, schema string) DataDestination
}

// DataDestination to write in the push process.
type DataDestination interface {
	Open(plan Plan, mode Mode, disableConstraints bool) *Error
	Commit() *Error
	RowWriter(table Table) (RowWriter, *Error)
	SetLogFolder(folderPath string)
	Close() *Error
}

// RowWriter write row to destination table
type RowWriter interface {
	// Write row in external datasource. where is optional and can contains additional key=value to use in the where clause.
	Write(row Row, where Row) *Error
}

type NoErrorCaptureRowWriter struct{}

func (necrw NoErrorCaptureRowWriter) Write(row Row, where Row) *Error {
	return &Error{"No error capture configured"}
}

// RowIterator iter over a collection of rows
type RowIterator interface {
	Next() bool
	Value() *Row
	Error() *Error
	Close() *Error
}

type Key struct {
	TableName  string
	ColumnName string
}

type Translator interface {
	FindValue(key Key, value Value) Value
	Load(keys []Key, rows RowIterator) *Error
}

type Observer interface {
	Pushed()
	Close()
}
