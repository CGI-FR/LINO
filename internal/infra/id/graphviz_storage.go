package id

import (
	"io/ioutil"
	"strconv"

	"github.com/awalterschulze/gographviz"
	"makeit.imfr.cgi.com/lino/pkg/id"
)

// DOTStorage provides storage in a graphviz DOT format
type DOTStorage struct{}

// NewDOTStorage create a new DOT storage
func NewDOTStorage() *DOTStorage {
	return &DOTStorage{}
}

// Store ingress descriptor in the DOT file
func (s *DOTStorage) Store(idef id.IngressDescriptor) *id.Error {
	graphName := strconv.Quote(idef.StartTable().Name())
	graph := gographviz.NewGraph()

	err := graph.SetName(graphName)
	if err != nil {
		return &id.Error{Description: err.Error()}
	}

	err = graph.SetDir(true)
	if err != nil {
		return &id.Error{Description: err.Error()}
	}

	for i := uint(0); i < idef.Relations().Len(); i++ {
		r := idef.Relations().Relation(i)
		src := strconv.Quote(r.Parent().Name())
		dst := strconv.Quote(r.Child().Name())

		err = graph.AddNode(graphName, src, nil)
		if err != nil {
			return &id.Error{Description: err.Error()}
		}

		err = graph.AddNode(graphName, dst, nil)
		if err != nil {
			return &id.Error{Description: err.Error()}
		}

		err = graph.AddEdge(src, dst, true, nil)
		if err != nil {
			return &id.Error{Description: err.Error()}
		}
	}

	err = ioutil.WriteFile("ingress-descriptor.dot", []byte(graph.String()), 0640)
	if err != nil {
		return &id.Error{Description: err.Error()}
	}

	return nil
}

func (s *DOTStorage) Read() (id.IngressDescriptor, *id.Error) {
	return nil, &id.Error{Description: "Not implemented"}
}
