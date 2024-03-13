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
	Export(ExportedRow) error
}

type DataSourceOption func(DataSource)

// DataSourceFactory exposes methods to create new datasources.
type DataSourceFactory interface {
	New(url string, schema string, options ...DataSourceOption) DataSource
}

// DataSource to read in the pull process.
type DataSource interface {
	Open() error
	RowReader(source Table, filter Filter) (RowReader, error)
	Read(source Table, filter Filter) (RowSet, error)
	Close() error
}

// RowReader over DataSource.
type RowReader interface {
	Next() bool
	Value() Row
	Error() error
	Close() error
}

// TraceListener receives diagnostic trace.
type TraceListener interface {
	TraceStep(Step) TraceListener
}

// NoTraceListener default implementation do nothing.
type NoTraceListener struct{}

// TraceStep catch Step event.
func (t NoTraceListener) TraceStep(s Step) TraceListener { return t }

type KeyStore interface {
	Has(row Row) bool
}
