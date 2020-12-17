package id

import "makeit.imfr.cgi.com/lino/pkg/id"

// TableStorage provides simple storage for one table Ingress Descriptor
type TableStorage struct {
	table id.Table
}

// NewTableStorage create a new TableStorage
func NewTableStorage(table id.Table) *TableStorage {
	return &TableStorage{
		table: table,
	}
}

// Store raise not implemented error
func (s *TableStorage) Store(adef id.IngressDescriptor) *id.Error {
	return &id.Error{Description: "Not implemented Store function for dynamique Ingress Descriptor Storage"}
}

// Read create new Ingress Descriptor with table as start table without relations
func (s *TableStorage) Read() (id.IngressDescriptor, *id.Error) {
	return id.NewIngressDescriptor(s.table, id.NewIngressRelationList([]id.IngressRelation{})), nil
}
