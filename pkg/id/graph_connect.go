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

func (g graph) findConnectedTables(start string) (TableList, *Error) {
	tables := []Table{}
	if err := g.visit(start, func(t Table) {
		tables = append(tables, t)
	}); err != nil {
		return nil, err
	}
	return NewTableList(tables), nil
}

func (g graph) getConnectedGraph(start string) (graph, *Error) {
	tables, err := g.findConnectedTables(start)
	if err != nil {
		return graph{}, err
	}
	return g.subGraph(tables), nil
}
