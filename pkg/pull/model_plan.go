package pull

import (
	"fmt"
	"strings"
)

type plan struct {
	filter Filter
	steps  StepList
}

// NewPlan initialize a new Plan object
func NewPlan(filter Filter, steps StepList) Plan {
	return plan{filter: filter, steps: steps}
}

func (p plan) InitFilter() Filter { return p.filter }
func (p plan) Steps() StepList    { return p.steps }
func (p plan) String() string {
	sb := &strings.Builder{}
	for i := uint(0); i < p.Steps().Len(); i++ {
		fmt.Fprintf(sb, "%v", p.Steps().Step(i))
	}
	return sb.String()
}
