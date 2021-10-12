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
	"encoding/json"

	"github.com/cgi-fr/jsonline/pkg/jsonline"
)

// Value is an untyped data.
type Value interface{}

// Row of data.
type Row map[string]Value

// Update Row with an other Row to generate a new one
func (r Row) Update(other Row) Row {
	for k, v := range other {
		r[k] = v
	}
	return r
}

// ExportableRow is a row but with keys ordered and values in export format for jsonline.
type ExportableRow interface {
	Has(key string) bool
	Get(key string) (interface{}, bool)
	GetOrNil(key string) interface{}
	Len() int
	Iter() func() (string, interface{}, bool)
	AsRow() Row
	json.Marshaler

	set(key string, value interface{})
	add(key string, others ...ExportableRow) bool
}

type row struct {
	jsonline.Row
}

// NewRow create a new Row
func NewExportableRow() ExportableRow {
	return &row{jsonline.NewRow()}
}

func (r *row) set(key string, value interface{}) {
	r.Row.Set(key, value)
}

func (r *row) add(key string, others ...ExportableRow) bool {
	value, has := r.Row.Get(key)

	if !has {
		value = []ExportableRow{}
	}

	array, ok := value.([]ExportableRow)
	if !ok {
		return false
	}

	array = append(array, others...)
	r.set(key, array)

	return true
}

func (r *row) GetOrNil(key string) interface{} {
	v, _ := r.Row.Get(key)

	return v
}

func (r *row) AsRow() Row {
	ex, _ := r.Row.Export()
	result := Row{}
	for k, v := range ex.(map[string]interface{}) {
		result[k] = v
	}
	return result
}
