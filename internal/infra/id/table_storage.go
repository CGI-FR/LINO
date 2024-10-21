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

import "github.com/cgi-fr/lino/pkg/id"

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
	return id.NewIngressDescriptor(s.table, []string{}, id.NewIngressRelationList([]id.IngressRelation{})), nil
}
