package pull

import (
	"fmt"
	"strings"
)

// NewCycle initialize a new Cycle object
func NewCycle(relations []Relation) Cycle {
	return relationList{uint(len(relations)), relations}
}

type cycleList struct {
	len   uint
	slice []Cycle
}

// NewCycleList initialize a new CycleList object
func NewCycleList(cycles []Cycle) CycleList {
	return cycleList{uint(len(cycles)), cycles}
}

func (l cycleList) Len() uint            { return l.len }
func (l cycleList) Cycle(idx uint) Cycle { return l.slice[idx] }
func (l cycleList) String() string {
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
