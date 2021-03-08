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

// Table from which to pull data.
type Table interface {
	Name() string
	PrimaryKey() []string
}

// Relation between two tables.
type Relation interface {
	Name() string
	Parent() Table
	Child() Table
	ParentKey() []string
	ChildKey() []string
	OppositeOf(tablename string) Table
}

// RelationList is a list of relations.
type RelationList interface {
	Len() uint
	Relation(idx uint) Relation
}

// Cycle is a list of relations.
type Cycle interface {
	RelationList
}

// CycleList is a list of cycles.
type CycleList interface {
	Len() uint
	Cycle(idx uint) Cycle
}

// Step group of follows to perform.
type Step interface {
	Index() uint
	Entry() Table
	Follow() Relation
	Relations() RelationList
	Cycles() CycleList
	NextSteps() StepList
}

// StepList list of steps to perform.
type StepList interface {
	Len() uint
	Step(uint) Step
}

// Plan of the puller process.
type Plan interface {
	InitFilter() Filter
	Steps() StepList
}

// Value is an untyped data.
type Value interface{}

// Filter applied to data tables.
type Filter interface {
	Limit() uint
	Values() Row
}

// Row of data.
type Row map[string]Value

// Update Row with an other Row to generate a new one
func (r Row) Update(other Row) Row {
	for k, v := range other {
		r[k] = v
	}
	return r
}

// Error is the error type returned by the domain
type Error struct {
	Description string
}

func (e *Error) Error() string {
	return e.Description
}
