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
	"container/list"
	"encoding/json"
	"fmt"
)

// Value is an untyped data.
type Value struct {
	Raw      interface{}
	Formated interface{}
	Export   bool
}

// Row of data.
type Row interface {
	Get(key string) Value
	GetIfHas(key string) (Value, bool)
	Has(key string) bool
	Set(key string, val Value)
	Delete(key string) (Value, bool)
	Export() map[string]interface{}
	Iter() func() (string, Value, bool)
	Len() int
	Update(other Row) Row
	json.Marshaler
}

type m map[string]Value

type row struct {
	m
	l    *list.List
	keys map[string]*list.Element
}

// NewRow create a new Row
func NewRow() Row {
	return &row{
		m:    make(map[string]Value),
		l:    list.New(),
		keys: make(map[string]*list.Element),
	}
}

func (r *row) Set(key string, value Value) {
	if _, ok := r.m[key]; !ok {
		r.keys[key] = r.l.PushBack(key)
	}
	r.m[key] = value
}

func (r *row) Has(key string) bool {
	_, ok := r.m[key]
	return ok
}

func (r *row) Get(key string) Value {
	return r.m[key]
}

func (r *row) GetIfHas(key string) (value Value, ok bool) {
	value, ok = r.m[key]
	return
}

func (r *row) Delete(key string) (value Value, ok bool) {
	value, ok = r.m[key]
	if ok {
		r.l.Remove(r.keys[key])
		delete(r.keys, key)
		delete(r.m, key)
	}
	return
}

func (r *row) Iter() func() (string, Value, bool) {
	e := r.l.Front()
	return func() (string, Value, bool) {
		if e != nil {
			key := e.Value.(string)
			e = e.Next()
			return key, r.m[key], true
		}
		return "", Value{}, false
	}
}

func (r *row) Len() int {
	return r.l.Len()
}

func (r *row) MarshalJSON() (res []byte, err error) {
	res = append(res, '{')
	front, back := r.l.Front(), r.l.Back()
	for e := front; e != nil; e = e.Next() {
		k := e.Value.(string)
		res = append(res, fmt.Sprintf("%q:", k)...)
		var b []byte
		b, err = json.Marshal(r.m[k].Formated)
		if err != nil {
			return
		}
		res = append(res, b...)
		if e != back {
			res = append(res, ',')
		}
	}
	res = append(res, '}')
	return
}

// Update Row with an other Row to generate a new one
func (r *row) Update(other Row) Row {
	iter := other.Iter()
	for k, v, ok := iter(); ok; k, v, ok = iter() {
		r.Set(k, v)
	}
	return r
}

func (r *row) Export() map[string]interface{} {
	result := map[string]interface{}{}
	for key, val := range r.m {
		if val.Export {
			if sr, ok := val.Formated.(Row); ok {
				result[key] = sr.Export()
			} else if sa, ok := val.Formated.([]Row); ok {
				array := make([]map[string]interface{}, 0, len(sa))
				for _, sar := range sa {
					array = append(array, sar.Export())
				}
				result[key] = array
			} else {
				result[key] = val.Formated
			}
		}
	}
	return result
}
