package id

import (
	"makeit.imfr.cgi.com/lino/pkg/id"
	"makeit.imfr.cgi.com/lino/pkg/relation"
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
