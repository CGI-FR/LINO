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

import (
	"fmt"
	"strings"
)

type idrelation struct {
	Relation
	lookUpParent bool
	lookUpChild  bool
	whereParent  string
	whereChild   string
	selectParent []string
	selectChild  []string
}

// NewIngressRelation initialize a new IngressRelation object
func NewIngressRelation(
	rel Relation,
	lookUpParent bool,
	lookUpChild bool,
	whereParent string,
	whereChild string,
	selectParent []string,
	selectChild []string,
) IngressRelation {
	return idrelation{
		Relation:     rel,
		lookUpParent: lookUpParent,
		lookUpChild:  lookUpChild,
		whereParent:  whereParent,
		whereChild:   whereChild,
		selectParent: selectParent,
		selectChild:  selectChild,
	}
}

func (r idrelation) Name() string           { return r.Relation.Name() }
func (r idrelation) Parent() Table          { return r.Relation.Parent() }
func (r idrelation) Child() Table           { return r.Relation.Child() }
func (r idrelation) LookUpParent() bool     { return r.lookUpParent }
func (r idrelation) WhereParent() string    { return r.whereParent }
func (r idrelation) SelectParent() []string { return r.selectParent }
func (r idrelation) LookUpChild() bool      { return r.lookUpChild }
func (r idrelation) WhereChild() string     { return r.whereChild }
func (r idrelation) SelectChild() []string  { return r.selectChild }
func (r idrelation) String() string {
	switch {
	case r.LookUpChild() && r.LookUpParent():
		return `↔` + r.Relation.String()
	case r.LookUpChild():
		return `→` + r.Relation.String()
	case r.LookUpParent():
		return `←` + r.Relation.String()
	}
	return r.Relation.String()
}

type idrelationList struct {
	len   uint
	slice []IngressRelation
	set   set
}

// NewIngressRelationList initialize a new IngressRelationList object
func NewIngressRelationList(relations []IngressRelation) IngressRelationList {
	set := newSet()
	for _, rel := range relations {
		set.add(rel.Name())
	}
	return idrelationList{uint(len(relations)), relations, set}
}

func (l idrelationList) Len() uint { return l.len }

func (l idrelationList) Relation(idx uint) IngressRelation { return l.slice[idx] }
func (l idrelationList) Contains(name string) bool         { return l.set.contains(name) }
func (l idrelationList) String() string {
	switch l.len {
	case 0:
		return ""
	case 1:
		return fmt.Sprint(l.slice[0])
	}
	sb := strings.Builder{}
	fmt.Fprintf(&sb, "%v", l.slice[0])
	for _, rel := range l.slice[1:] {
		fmt.Fprintf(&sb, " %v", rel)
	}
	return sb.String()
}

// NewIngressDescriptor initialize a new IngressDescriptor object
func NewIngressDescriptor(start Table, selectColumns []string, relations IngressRelationList) IngressDescriptor {
	return id{startTable: table{name: start.Name()}, selectColumns: selectColumns, relations: relations}
}

type id struct {
	startTable    table
	selectColumns []string
	relations     IngressRelationList
}

func (id id) StartTable() Table              { return id.startTable }
func (id id) Select() []string               { return id.selectColumns }
func (id id) Relations() IngressRelationList { return id.relations }
func (id id) String() string {
	return fmt.Sprintf("%v [%v] (%v)", id.startTable, id.selectColumns, id.relations)
}
