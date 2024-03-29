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
	"fmt"

	"github.com/cgi-fr/rimo/pkg/model"
	"github.com/cgi-fr/rimo/pkg/rimo"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
)

type Config struct {
	SampleSize     uint
	Distinct       bool
	Limit          uint
	Tables         []string
	Wheres         map[string]string   // tablename -> where clause
	ExcludeColumns map[string][]string // tablename -> list of columns
}

type Driver struct {
	analyser rimo.Driver
	ds       DataSource
	exf      ExtractorFactory
	w        Writer
	cfg      Config

	tables  []string
	columns []string

	curTable  int
	curColumn int
}

func NewDriver(datasource DataSource, exf ExtractorFactory, writer Writer, cfg Config) *Driver {
	tables := datasource.ListTables()
	if len(cfg.Tables) > 0 {
		tables = cfg.Tables
	}

	return &Driver{
		analyser:  rimo.Driver{SampleSize: cfg.SampleSize, Distinct: cfg.Distinct},
		ds:        datasource,
		exf:       exf,
		w:         writer,
		cfg:       cfg,
		tables:    tables,
		columns:   []string{},
		curTable:  -1,
		curColumn: -1,
	}
}

func (d *Driver) Open() error  { return nil }
func (d *Driver) Close() error { return nil }

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

		log.Debug().Msg("go to next column")

		return true
	}

	if d.curTable+1 == len(d.tables) {
		log.Debug().Msg("last column of last table reached")

		return false
	}

	// no more columns, check if there is more tables
	for d.curTable+1 < len(d.tables) {
		// yes, increase table index and read columns
		d.curTable++
		d.curColumn = 0
		d.columns = d.ds.ListColumn(d.tables[d.curTable])

		d.columns = slices.DeleteFunc(d.columns, func(column string) bool {
			if excluded, ok := d.cfg.ExcludeColumns[d.tables[d.curTable]]; ok {
				return slices.Contains(excluded, column)
			}
			return false
		})

		// should we try next table because there is no column in this table
		if len(d.columns) > 0 {
			log.Debug().
				Str("table", d.tables[d.curTable]).
				Strs("columns", d.columns).
				Msg("next table")

			break // table has columns, let's go!
		}
	}

	// last table is not passed
	return d.curTable < len(d.tables) && len(d.columns) > 0
}

func (d *Driver) Col() (rimo.ColReader, error) { //nolint:ireturn
	where := d.cfg.Wheres[d.tables[d.curTable]]

	return &ValueIterator{
		Extractor: d.exf.New(d.tables[d.curTable], d.columns[d.curColumn], d.cfg.Limit, where),
		tableName: d.tables[d.curTable],
		colName:   d.columns[d.curColumn],
		nextValue: nil,
		err:       nil,
	}, nil
}

func (d *Driver) Export(base *model.Base) error {
	if err := d.w.Write(base); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

type ValueIterator struct {
	Extractor
	tableName string
	colName   string
	nextValue interface{}
	err       error
}

func (vi *ValueIterator) Open() error       { return vi.Extractor.Open() }  //nolint:wrapcheck
func (vi *ValueIterator) Close() error      { return vi.Extractor.Close() } //nolint:wrapcheck
func (vi *ValueIterator) ColName() string   { return vi.colName }
func (vi *ValueIterator) TableName() string { return vi.tableName }

func (vi *ValueIterator) Next() bool {
	var result bool

	result, vi.nextValue, vi.err = vi.ExtractValue()

	return result
}

func (vi *ValueIterator) Value() (any, error) {
	if vi.err != nil {
		return nil, fmt.Errorf("could not extract value: %w", vi.err)
	}

	return vi.nextValue, nil
}
