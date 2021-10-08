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

// Row of data.
type Row interface {
	Has(key string) bool
	Get(key string) interface{}
	Set(key string, value interface{})
	SetValue(key string, value jsonline.Value)
	SetRow(key string, row Row)
	AddRows(key string, rows ...Row) bool
	Update(other Row) Row
	Len() int
	Iter() func() (string, interface{}, bool)
	json.Marshaler
}

type row struct {
	wrapped jsonline.Row
}

// NewRow create a new Row
func NewRow() Row {
	return &row{wrapped: jsonline.NewRow()}
}

func (r *row) Has(key string) bool {
	return r.wrapped.Get(key) != nil
}

func (r *row) Get(key string) interface{} {
	return r.wrapped.Get(key).Raw()
}

func (r *row) Set(key string, value interface{}) {
	if r.Has(key) {
		r.wrapped.Get(key).Import(value)
	} else {
		r.wrapped.Set(key, jsonline.NewValueAuto(value))
	}
}

func (r *row) SetValue(key string, value jsonline.Value) {
	r.wrapped.Set(key, value)
}

func (r *row) SetRow(key string, row Row) {
	r.Set(key, row)
}

func (r *row) AddRows(key string, rows ...Row) bool {
	value := r.wrapped.Get(key)
	if value == nil {
		r.wrapped.Set(key, jsonline.NewValueAuto([]Row{}))
	}

	array, ok := r.wrapped.Get(key).Raw().([]Row)
	if !ok {
		return false
	}

	array = append(array, rows...)
	r.Set(key, array)

	return true
}

func (r *row) Update(other Row) Row {
	iter := other.Iter()
	for k, v, ok := iter(); ok; k, v, ok = iter() {
		r.Set(k, v)
	}
	return r
}

func (r *row) Len() int {
	m := r.wrapped.Raw().(map[string]interface{})
	return len(m)
}

func (r *row) Iter() func() (string, interface{}, bool) {
	iter := r.wrapped.Iter()

	return func() (string, interface{}, bool) {
		key, val, ok := iter()
		if ok {
			return key, val.Raw(), ok
		}
		return key, nil, ok
	}
}

func (r *row) MarshalJSON() ([]byte, error) {
	return r.wrapped.MarshalJSON()
}
