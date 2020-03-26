package id

import (
	"fmt"
	"strings"
)

type step struct {
	index     uint
	entry     Table
	following IngressRelation
	relations IngressRelationList
	tables    TableList
	cycles    CycleList
	previous  uint
}

// NewStep initialize a new Step object
func NewStep(index uint, entry Table, following IngressRelation, relations IngressRelationList, tables TableList, cycles CycleList, previousStep uint) Step {
	return step{index: index, entry: entry, following: following, relations: relations, tables: tables, cycles: cycles, previous: previousStep}
}
func (s step) Index() uint                    { return s.index }
func (s step) Entry() Table                   { return s.entry }
func (s step) Following() IngressRelation     { return s.following }
func (s step) Relations() IngressRelationList { return s.relations }
func (s step) Tables() TableList              { return s.tables }
func (s step) Cycles() CycleList              { return s.cycles }
func (s step) PreviousStep() uint             { return s.previous }
func (s step) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "step %v - pull rows from %v", s.Index(), s.Entry())
	if s.PreviousStep() != 0 {
		fmt.Fprintf(sb, " following %v relationship for rows pulled at step %v", s.Following(), s.PreviousStep())
	}
	for idx := uint(0); idx < s.Cycles().Len(); idx++ {
		cycle := s.Cycles().Cycle(idx)
		if cycle.Len() == 1 && cycle.Relation(0).LookUpChild() && cycle.Relation(0).LookUpParent() {
			fmt.Fprintf(sb, ", then follow %v relationship (round trip)", cycle)
		} else {
			fmt.Fprintf(sb, ", then follow %v relationships (loop until data exhaustion)", cycle)
		}
	}
	return sb.String()
}

type pullionPlan struct {
	len       uint
	slice     []Step
	relations IngressRelationList
	tables    TableList
}

// NewPullionPlan initialize a new PullionPlan object
func NewPullionPlan(steps []Step, relations IngressRelationList, tables TableList) PullionPlan {
	return pullionPlan{uint(len(steps)), steps, relations, tables}
}

func (l pullionPlan) Len() uint                      { return l.len }
func (l pullionPlan) Step(idx uint) Step             { return l.slice[idx] }
func (l pullionPlan) Relations() IngressRelationList { return l.relations }
func (l pullionPlan) Tables() TableList              { return l.tables }
func (l pullionPlan) String() string {
	switch l.len {
	case 0:
		return ""
	case 1:
		return fmt.Sprint(l.slice[0])
	}
	sb := strings.Builder{}
	fmt.Fprintf(&sb, "%v", l.slice[0])
	for _, rel := range l.slice[1:] {
		fmt.Fprintf(&sb, ", then %v", rel)
	}
	return sb.String()
}
