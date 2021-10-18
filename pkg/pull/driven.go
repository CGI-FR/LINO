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

// RowExporter receives pulled rows one by one.
type RowExporter interface {
	Export(Row) *Error
}

// DataSourceFactory exposes methods to create new datasources.
type DataSourceFactory interface {
	New(url string, schema string) DataSource
}

// DataSource to read in the pull process.
type DataSource interface {
	Open() *Error
	RowReader(source Table, filter Filter) (RowReader, *Error)
	Close() *Error
}

// RowReader over DataSource.
type RowReader interface {
	Next() bool
	Value() Row
	Error() *Error
}

// OneOneEmptyRowReader return one empty row
type OneEmptyRowReader struct {
	done bool
}

func NewOneEmptyRowReader() *OneEmptyRowReader {
	return &OneEmptyRowReader{false}
}

// Next is always false except for the first one
func (r *OneEmptyRowReader) Next() bool {
	result := !r.done
	r.done = true
	return result
}

// Value is always an empty row
func (r OneEmptyRowReader) Value() Row { return Row{} }

// Error return always nil
func (r OneEmptyRowReader) Error() *Error { return nil }

// TraceListener receives diagnostic trace
type TraceListener interface {
	TraceStep(Step, Filter) TraceListener
}

// NoTraceListener default implementation do nothing.
type NoTraceListener struct{}

// TraceStep catch Step event.
func (t NoTraceListener) TraceStep(s Step, filter Filter) TraceListener { return t }
