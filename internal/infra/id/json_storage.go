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
	"encoding/json"
	"os"

	"github.com/cgi-fr/lino/pkg/id"
)

// JSONStructure of the file.
type JSONStructure struct {
	Version           string                `json:"version"`
	IngressDescriptor JSONIngressDescriptor `json:"IngressDescriptor"`
}

// JSONIngressDescriptor defines how to store an ingress descriptor in JSON format.
type JSONIngressDescriptor struct {
	StartTable string         `json:"startTable"`
	Relations  []JSONRelation `json:"relations"`
}

// JSONRelation defines how to store a relation in JSON format.
type JSONRelation struct {
	Name   string    `json:"name"`
	Parent JSONTable `json:"parent"`
	Child  JSONTable `json:"child"`
}

// JSONTable defines how to store a table in JSON format.
type JSONTable struct {
	Name   string `json:"name"`
	Lookup bool   `json:"lookup"`
}

// JSONStorage provides storage in a local JSON file
type JSONStorage struct {
	file os.File
}

// NewJSONStorage create a new JSON storage
func NewJSONStorage(file os.File) *JSONStorage {
	return &JSONStorage{file}
}

// Store ingress descriptor in the JSON file
func (s *JSONStorage) Store(id id.IngressDescriptor) *id.Error {
	structure := JSONStructure{
		Version: Version,
	}

	relations := []JSONRelation{}
	list := id.Relations()
	for i := uint(0); i < list.Len(); i++ {
		relation := list.Relation(i)
		relations = append(relations, JSONRelation{
			Name:   relation.Name(),
			Parent: JSONTable{Name: relation.Parent().Name(), Lookup: relation.LookUpParent()},
			Child:  JSONTable{Name: relation.Child().Name(), Lookup: relation.LookUpChild()},
		})
	}

	structure.IngressDescriptor = JSONIngressDescriptor{
		StartTable: id.StartTable().Name(),
		Relations:  relations,
	}

	err := writeJSONFile(&structure, s.file)
	if err != nil {
		return err
	}

	return nil
}

func (s *JSONStorage) Read() (id.IngressDescriptor, *id.Error) {
	return nil, &id.Error{Description: "not implemented"}
}

func writeJSONFile(structure *JSONStructure, file os.File) *id.Error {
	out, err := json.MarshalIndent(structure, "", " ")
	if err != nil {
		return &id.Error{Description: err.Error()}
	}

	_, err = file.Write(out)
	if err != nil {
		return &id.Error{Description: err.Error()}
	}

	return nil
}
