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

// stack is a simple implementation of the stack pattern
type stack struct {
	items []Table
}

func newStack() stack {
	return stack{items: []Table{}}
}

// push an element on top of the stack
func (s *stack) push(v Table) {
	s.items = append(s.items, v)
}

// peek the element on top of the stack (return nil is stack is empty)
/* func (s *stack) peek() *Table {
	var v *Table
	if len(s.vertices) > 0 {
		v = s.vertices[len(s.vertices)-1]
	}
	return v
} */

// pop the top element (return nil is stack is empty)
func (s *stack) pop() Table {
	var v Table
	if len(s.items) > 0 {
		v = s.items[len(s.items)-1]
		s.items = s.items[:len(s.items)-1]
	}
	return v
}

// empty is true if the stack is empty
/* func (s *stack) empty() bool {
	return len(s.vertices) == 0
} */

func (s *stack) asSlice() []Table {
	result := []Table{}
	result = append(result, s.items...)
	return result
}
