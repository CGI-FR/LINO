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
	"fmt"
	"strings"
)

type relation struct {
	name      string
	parent    Table
	child     Table
	parentKey []string
	childKey  []string
}

// NewRelation initialize a new Relation object
func NewRelation(name string, parent Table, child Table, parentKey []string, childKey []string) Relation {
	return relation{name: name, parent: parent, child: child, parentKey: parentKey, childKey: childKey}
}

func (r relation) Name() string        { return r.name }
func (r relation) Parent() Table       { return r.parent }
func (r relation) Child() Table        { return r.child }
func (r relation) ParentKey() []string { return r.parentKey }
func (r relation) ChildKey() []string  { return r.childKey }
func (r relation) String() string      { return r.name }
func (r relation) OppositeOf(tablename string) Table {
	if r.Child().Name() == tablename {
		return r.Parent()
	}
	return r.Child()
}

type relationList struct {
	len   uint
	slice []Relation
}

// NewRelationList initialize a new RelationList object
func NewRelationList(relations []Relation) RelationList {
	return relationList{uint(len(relations)), relations}
}

func (l relationList) Len() uint                  { return l.len }
func (l relationList) Relation(idx uint) Relation { return l.slice[idx] }
func (l relationList) String() string {
	switch l.len {
	case 0:
		return ""
	case 1:
		return fmt.Sprint(l.slice[0])
	}
	sb := strings.Builder{}
	fmt.Fprintf(&sb, "%v", l.slice[0])
	for _, rel := range l.slice[1:] {
		fmt.Fprintf(&sb, " -> %v", rel)
	}
	return sb.String()
}
