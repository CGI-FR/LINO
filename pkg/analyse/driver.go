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

import (
	"github.com/cgi-fr/rimo/pkg/model"
	"github.com/cgi-fr/rimo/pkg/rimo"
)

type Driver struct {
	analyser rimo.Driver
	ds       DataSource
	exf      ExtractorFactory

	tables  []string
	columns []string

	curTable  int
	curColumn int
}

func NewDriver(ds DataSource) *Driver {
	return &Driver{
		analyser:  rimo.Driver{SampleSize: 5, Distinct: false}, //nolint:gomnd
		ds:        ds,
		tables:    ds.ListTables(),
		columns:   []string{},
		curTable:  -1,
		curColumn: -1,
	}
}

// Analyse performs statistics on datasource.
func (d *Driver) Analyse() error {
	return d.analyser.AnalyseBase(d, d) //nolint:wrapcheck
}

func (d *Driver) BaseName() string {
	return d.ds.Name()
}

// Next returns next column in database
func (d *Driver) Next() bool {
	// check if there is more columns in current table
	if d.curColumn+1 < len(d.columns) {
		// yes, so increase column index
		d.curColumn++

		return true
	}

	// no more columns, check if there is more tables
	for d.curTable+1 < len(d.tables) {
		// yes, increase table index and read columns
		d.curTable++
		d.curColumn = 0
		d.columns = d.ds.ListColumn(d.tables[d.curTable])

		// should we try next table because there is no column in this table
		if len(d.columns) > 0 {
			break // table has columns, let's go!
		}
	}

	// last table is not passed
	return d.curTable < len(d.tables)
}

func (d *Driver) Col() (rimo.ColReader, error) { //nolint:ireturn
	return &ValueIterator{
		Extractor: d.exf.New(d.tables[d.curTable], d.columns[d.curColumn]),
		tableName: d.tables[d.curTable],
		colName:   d.columns[d.curColumn],
	}, nil
}

func (d *Driver) Export(base *model.Base) error {
	return nil
}

type ValueIterator struct {
	Extractor
	tableName string
	colName   string
}

func (vi *ValueIterator) ColName() string     { return vi.colName }
func (vi *ValueIterator) TableName() string   { return vi.tableName }
func (vi *ValueIterator) Next() bool          { panic("") }
func (vi *ValueIterator) Value() (any, error) { panic("") }

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
	values, err := ci.ExtractValues(ci.tables[0], ci.column[0])
	return values, ci.column[0], ci.tables[0], err
}
