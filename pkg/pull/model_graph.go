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

func (g Graph) addMissingColumns(t Table) Table {
	if len(g.Relations[t.Name]) > 0 {
		g.Cached[t.Name] = true
	}

	if len(t.Columns) > 0 {
		for _, relation := range g.Relations[t.Name] {
			t.addMissingColumns(relation.Local.Keys...)
		}
	}

	return t
}
