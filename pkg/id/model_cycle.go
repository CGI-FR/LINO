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
