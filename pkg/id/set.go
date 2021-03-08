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

// set is a simple implementation of the set pattern.
type set map[string]string

func newSet() set {
	return map[string]string{}
}

// add an element in the set
func (s set) add(v string) {
	s[v] = v
}

// remove an element in the set
func (s set) remove(v string) {
	delete(s, v)
}

// contains the element ?
func (s set) contains(v string) bool {
	_, ok := s[v]
	return ok
}
