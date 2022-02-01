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

import "sort"

type plan struct {
	firstTable Table
	relations  []Relation
}

// NewPlan initialize a new Plan object
func NewPlan(first Table, relations []Relation) Plan {
	return plan{firstTable: first, relations: relations}
}

func (p plan) FirstTable() Table { return p.firstTable }
func (p plan) RelationsFromTable(table Table) map[string]Relation {
	result := map[string]Relation{}
	for _, r := range p.relations {
		if r.Parent().Name() == table.Name() {
			result[r.Name()] = r
		} else if r.Child().Name() == table.Name() {
			result[r.Name()] = r
		}
	}
	return result
}

func (p plan) Tables() []Table {
	tableOrder := map[string]int{}

	name2Table := map[string]Table{}
	for _, r := range p.relations {
		tableOrder[r.Child().Name()] = tableOrder[r.Parent().Name()] + 1
		name2Table[r.Child().Name()] = r.Child()
		name2Table[r.Parent().Name()] = r.Parent()
	}

	// propagate priority to children
	for i := 0; i < len(tableOrder); i++ {
		for _, r := range p.relations {
			tableOrder[r.Child().Name()] = tableOrder[r.Parent().Name()] + 1
		}
	}

	type to struct {
		t Table
		o int
	}

	tables := []to{}

	for name, table := range name2Table {
		tables = append(tables, to{table, tableOrder[name]})
	}

	sort.Slice(tables, func(i, j int) bool {
		return tables[i].o > tables[j].o
	})

	result := []Table{}
	for _, v := range tables {
		result = append(result, v.t)
	}
	return result
}
