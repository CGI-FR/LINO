package id

import (
	"fmt"
	"strings"
)

type idrelation struct {
	Relation
	lookUpParent bool
	lookUpChild  bool
}

// NewIngressRelation initialize a new IngressRelation object
func NewIngressRelation(rel Relation, lookUpParent bool, lookUpChild bool) IngressRelation {
	return idrelation{Relation: rel, lookUpParent: lookUpParent, lookUpChild: lookUpChild}
}

func (r idrelation) Name() string       { return r.Relation.Name() }
func (r idrelation) Parent() Table      { return r.Relation.Parent() }
func (r idrelation) Child() Table       { return r.Relation.Child() }
func (r idrelation) LookUpParent() bool { return r.lookUpParent }
func (r idrelation) LookUpChild() bool  { return r.lookUpChild }
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
func NewIngressDescriptor(start Table, relations IngressRelationList) IngressDescriptor {
	return id{startTable: table{name: start.Name()}, relations: relations}
}

type id struct {
	startTable table
	relations  IngressRelationList
}

func (id id) StartTable() Table              { return id.startTable }
func (id id) Relations() IngressRelationList { return id.relations }
func (id id) String() string                 { return fmt.Sprintf("%v (%v)", id.startTable, id.relations) }
