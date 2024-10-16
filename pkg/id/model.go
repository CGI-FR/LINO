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

package id

// Table involved in an puller plan.
type Table interface {
	Name() string
	String() string
}

// TableList involved in an puller plan.
type TableList interface {
	Len() uint
	Table(idx uint) Table
	Contains(string) bool
	String() string
}

// Relation involved in an puller plan.
type Relation interface {
	Name() string
	Parent() Table
	Child() Table
	String() string
}

// RelationList involved in an puller plan.
type RelationList interface {
	Len() uint
	Relation(idx uint) Relation
	Contains(string) bool
	String() string
}

// IngressRelation describe how a relation will be accessed.
type IngressRelation interface {
	Relation
	LookUpChild() bool
	WhereChild() string
	SelectChild() []string
	LookUpParent() bool
	WhereParent() string
	SelectParent() []string
}

// IngressRelationList involved in an puller plan.
type IngressRelationList interface {
	Len() uint
	Relation(idx uint) IngressRelation
	Contains(string) bool
	String() string
}

// IngressDescriptor from which the puller plan will be computed.
type IngressDescriptor interface {
	StartTable() Table
	Select() []string
	Relations() IngressRelationList
	String() string
}

// A Cycle in the puller plan.
type Cycle interface {
	IngressRelationList
}

// A CycleList in the puller plan.
type CycleList interface {
	Len() uint
	Cycle(idx uint) Cycle
	String() string
}

// An Step gives required information to pull data.
type Step interface {
	Index() uint
	Entry() Table
	Following() IngressRelation
	Relations() IngressRelationList
	Tables() TableList
	Cycles() CycleList
	PreviousStep() uint
	String() string
}

// PullerPlan is the computed plan that lists all steps required to pull data.
type PullerPlan interface {
	Len() uint
	Step(idx uint) Step
	Relations() IngressRelationList
	Tables() TableList
	String() string
	Select() []string
}

// Error is the error type returned by the domain
type Error struct {
	Description string
}

func (e *Error) Error() string {
	return e.Description
}
