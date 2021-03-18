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
