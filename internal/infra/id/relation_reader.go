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
	"github.com/cgi-fr/lino/pkg/id"
	"github.com/cgi-fr/lino/pkg/relation"
)

// RelationReader is an adapter to read relations from the relation domain.
type RelationReader struct {
	relations []relation.Relation
}

// NewRelationReader create a new relations reader
func NewRelationReader(relations []relation.Relation) *RelationReader {
	return &RelationReader{relations: relations}
}

func (r *RelationReader) Read() (id.RelationList, *id.Error) {
	result := []id.Relation{}
	for _, relation := range r.relations {
		result = append(result, id.NewRelation(relation.Name, id.NewTable(relation.Parent.Name), id.NewTable(relation.Child.Name)))
	}
	return id.NewRelationList(result), nil
}
