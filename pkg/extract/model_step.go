package extract

import (
	"fmt"
	"strings"
)

type step struct {
	index     uint
	entry     Table
	follow    Relation
	relations RelationList
	cycles    CycleList
	nextSteps StepList
}

// NewStep initialize a new Step object
func NewStep(index uint, entry Table, follow Relation, relations RelationList, cycles CycleList, nextSteps StepList) Step {
	return step{index: index, entry: entry, follow: follow, relations: relations, cycles: cycles, nextSteps: nextSteps}
}

func (s step) Index() uint             { return s.index }
func (s step) Entry() Table            { return s.entry }
func (s step) Follow() Relation        { return s.follow }
func (s step) Relations() RelationList { return s.relations }
func (s step) Cycles() CycleList       { return s.cycles }
func (s step) NextSteps() StepList     { return s.nextSteps }
func (s step) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "step %v - extract rows from %v", s.Index(), s.Entry())
	if s.Follow() != nil {
		fmt.Fprintf(sb, " following %v relationship", s.Follow())
	}
	if s.NextSteps().Len() > 0 {
		fmt.Fprintf(sb, " then execute step(s) :")
	}
	for idx := uint(0); idx < s.NextSteps().Len(); idx++ {
		fmt.Fprintf(sb, " %v", s.NextSteps().Step(idx).Index())
	}
	return sb.String()
}

type stepList struct {
	len   uint
	slice []Step
}

// NewStepList initialize a new StepList object
func NewStepList(steps []Step) StepList {
	return stepList{uint(len(steps)), steps}
}

func (l stepList) Len() uint          { return l.len }
func (l stepList) Step(idx uint) Step { return l.slice[idx] }
func (l stepList) String() string {
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
