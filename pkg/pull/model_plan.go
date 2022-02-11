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

func (plan Plan) buildGraph() Graph {
	relations := map[TableName]RelationSet{}

	for _, relation := range plan.Relations {
		relations[relation.Local.Table.Name] = append(relations[relation.Local.Table.Name], relation)
	}

	relationsWithMissingColumns := map[TableName]RelationSet{}
	cached := map[TableName]bool{}

	for _, relation := range plan.Relations {
		relation.Local.Table.addMissingColumns(relation.Local.Keys...)

		cached[relation.Local.Table.Name] = true
		if len(relations[relation.Foreign.Table.Name]) > 0 {
			cached[relation.Foreign.Table.Name] = true
		}

		if len(relation.Foreign.Table.Columns) > 0 {
			for _, follow := range relations[relation.Foreign.Table.Name] {
				relation.Foreign.Table.addMissingColumns(follow.Local.Keys...)
			}
		}

		relationsWithMissingColumns[relation.Local.Table.Name] =
			append(relationsWithMissingColumns[relation.Local.Table.Name], relation)
	}

	return Graph{Relations: relationsWithMissingColumns, Components: plan.Components, Cached: cached}
}
