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

package push

type relation struct {
	name   string
	parent Table
	child  Table
}

// NewRelation initialize a new Relation object
func NewRelation(name string, parent Table, child Table) Relation {
	return relation{name: name, parent: parent, child: child}
}

func (r relation) Name() string   { return r.name }
func (r relation) Parent() Table  { return r.parent }
func (r relation) Child() Table   { return r.child }
func (r relation) String() string { return r.name }
func (r relation) OppositeOf(table Table) Table {
	if r.Child().Name() == table.Name() {
		return r.Parent()
	}
	return r.Child()
}
