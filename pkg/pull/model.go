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
	"github.com/cgi-fr/jsonline/pkg/jsonline"
)

type Cardinality bool

const (
	Many Cardinality = true
	One  Cardinality = false
)

type (
	TableName    string
	RelationName string
)

type Column struct {
	Name   string
	Export string
}

type Table struct {
	Name    TableName
	Keys    []string
	Columns []Column

	template jsonline.Template
}

type RelationTip struct {
	Table Table
	Keys  []string
}

type Relation struct {
	Name        RelationName
	Cardinality Cardinality
	Local       RelationTip
	Foreign     RelationTip
}

type RelationSet []Relation

type Plan struct {
	Relations  RelationSet
	Components map[TableName]uint // <= could be deduced from relations with tarjan algorithm
}

type Graph struct {
	Relations  map[TableName]RelationSet
	Components map[TableName]uint
	Cached     map[TableName]bool
}

type Row map[string]interface{}

type RowSet []Row

type DataSet map[TableName]RowSet

type Filter struct {
	Limit    uint
	Values   Row
	Where    string
	Distinct bool
}

// ExportedRow is a row but with keys ordered and values in export format for jsonline.
type ExportedRow struct {
	jsonline.Row
}

func (er ExportedRow) GetOrNil(key string) interface{} {
	v, _ := er.Get(key)

	return v
}
