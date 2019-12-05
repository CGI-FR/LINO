package id

import (
	"fmt"
	"strings"
)

type relation struct {
	name   string
	parent Table
	child  Table
}

// NewRelation initialize a new Relation object
func NewRelation(name string, parent Table, child Table) Relation {
	return relation{name: name, parent: parent, child: child}
}

func (r relation) Name() string   { return r.name }
func (r relation) Parent() Table  { return r.parent }
func (r relation) Child() Table   { return r.child }
func (r relation) String() string { return r.name }

type relationList struct {
	len   uint
	slice []Relation
	set   set
}

// NewRelationList initialize a new RelationList object
func NewRelationList(relations []Relation) RelationList {
	set := newSet()
	for _, rel := range relations {
		set.add(rel.Name())
	}
	return relationList{uint(len(relations)), relations, set}
}

func (l relationList) Len() uint                  { return l.len }
func (l relationList) Relation(idx uint) Relation { return l.slice[idx] }
func (l relationList) Contains(name string) bool  { return l.set.contains(name) }
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
		fmt.Fprintf(&sb, ", %v", rel)
	}
	return sb.String()
}
