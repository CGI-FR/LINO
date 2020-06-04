package id

import (
	"fmt"
	"strings"
)

type cycleList struct {
	len   uint
	slice []IngressRelationList
}

// NewCycleList initialize a new CycleList object
func NewCycleList(cycles []IngressRelationList) CycleList {
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
	fmt.Fprintf(&sb, "[%v", l.slice[0])
	for _, rel := range l.slice[1:] {
		fmt.Fprintf(&sb, ", %v", rel)
	}
	fmt.Fprintf(&sb, "]")
	return sb.String()
}
