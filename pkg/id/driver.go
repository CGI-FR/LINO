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
	"fmt"

	"github.com/rs/zerolog/log"
)

// Create and store ingress descriptor for the given start table and relation set.
func Create(startTable string, relReader RelationReader, storage Storage) *Error {
	relations, err := relReader.Read()
	if err != nil {
		return err
	}

	ingressRels := []IngressRelation{}
	for i := uint(0); i < relations.Len(); i++ {
		rel := relations.Relation(i)
		ingressRels = append(ingressRels, NewIngressRelation(rel, false, false, "", ""))
	}

	fullGraph := newGraph(NewIngressRelationList(ingressRels))

	connectedGraph, err := fullGraph.getConnectedGraph(startTable)
	if err != nil {
		return err
	}

	setLookUpChild := newSet()
	if err := connectedGraph.visitChildren(startTable, func(t Table) {
		outgoingRelations := connectedGraph.relationsFrom(t)
		for _, rel := range outgoingRelations {
			setLookUpChild.add(rel.Name())
		}
	}); err != nil {
		return err
	}

	adrelations := []IngressRelation{}
	for i := uint(0); i < connectedGraph.relations.Len(); i++ {
		rel := connectedGraph.relations.Relation(i)
		if setLookUpChild.contains(rel.Name()) {
			adrelations = append(adrelations, NewIngressRelation(NewRelation(rel.Name(), rel.Parent(), rel.Child()), false, true, rel.WhereParent(), rel.WhereChild()))
		} else {
			adrelations = append(adrelations, NewIngressRelation(NewRelation(rel.Name(), rel.Parent(), rel.Child()), false, false, rel.WhereParent(), rel.WhereChild()))
		}
	}

	id := NewIngressDescriptor(NewTable(startTable), NewIngressRelationList(adrelations))

	err = storage.Store(id)
	if err != nil {
		return err
	}

	return nil
}

// SetStartTable update ingress descriptor start table
func SetStartTable(table Table, storage Storage) *Error {
	id, err := storage.Read()
	if err != nil {
		return err
	}

	tableExist := false
	for i := uint(0); i < id.Relations().Len(); i++ {
		rel := id.Relations().Relation(i)
		tableExist = tableExist || rel.Parent() == table || rel.Child() == table
	}
	if !tableExist {
		return &Error{Description: fmt.Sprintf("Table %s doesn't exist", table.Name())}
	}

	updatedID := NewIngressDescriptor(table, id.Relations())

	err = storage.Store(updatedID)
	if err != nil {
		return err
	}
	return nil
}

// SetChildLookup update child lookup relation's parameter in ingress descriptor
func SetChildLookup(relation string, flag bool, storage Storage) *Error {
	id, err := storage.Read()
	if err != nil {
		return err
	}

	if !id.Relations().Contains(relation) {
		return &Error{Description: fmt.Sprintf("Relation %s doesn't exist", relation)}
	}

	relations := make([]IngressRelation, id.Relations().Len())

	for i := uint(0); i < id.Relations().Len(); i++ {
		rel := id.Relations().Relation(i)
		if rel.Name() == relation {
			rel = NewIngressRelation(NewRelation(rel.Name(), rel.Parent(), rel.Child()), rel.LookUpParent(), flag, rel.WhereParent(), rel.WhereChild())
		}
		relations[i] = rel
	}

	updatedID := NewIngressDescriptor(id.StartTable(), NewIngressRelationList(relations))

	err = storage.Store(updatedID)
	if err != nil {
		return err
	}
	return nil
}

// SetParentLookup update parent lookup relation's parameter in ingress descriptor
func SetParentLookup(relation string, flag bool, storage Storage) *Error {
	id, err := storage.Read()
	if err != nil {
		return err
	}

	if !id.Relations().Contains(relation) {
		return &Error{Description: fmt.Sprintf("Relation %s doesn't exist", relation)}
	}

	relations := make([]IngressRelation, id.Relations().Len())

	for i := uint(0); i < id.Relations().Len(); i++ {
		rel := id.Relations().Relation(i)
		if rel.Name() == relation {
			rel = NewIngressRelation(NewRelation(rel.Name(), rel.Parent(), rel.Child()), flag, rel.LookUpChild(), rel.WhereParent(), rel.WhereChild())
		}
		relations[i] = rel
	}

	updatedID := NewIngressDescriptor(id.StartTable(), NewIngressRelationList(relations))

	err = storage.Store(updatedID)
	if err != nil {
		return err
	}
	return nil
}

// GetPullerPlan returns the calculated puller plan.
func GetPullerPlan(storage Storage) (PullerPlan, *Error) {
	id, err := storage.Read()
	if err != nil {
		return nil, err
	}

	g := newGraph(id.Relations())
	g = g.slim() // remove inactive relations

	var startComponent TableList
	components := g.condense()
	for i, component := range components {
		cycles := g.subGraph(component).relCycles(id.StartTable())
		log.Debug().Msg(fmt.Sprintf("component %v - %v - %v", i, component, cycles))
		if component.Contains(id.StartTable().Name()) {
			startComponent = component
		}
	}
	log.Debug().Msg("")

	startRelationsList := NewIngressRelationList([]IngressRelation{})
	startTableList := NewTableList([]Table{id.StartTable()})
	startCycles := NewCycleList([]IngressRelationList{})
	if startComponent != nil {
		sg := g.subGraph(startComponent)
		startCycles = sg.relCycles(id.StartTable())
		startTableList = startComponent
		startRelationsList = sg.relations
	}
	steps := []Step{
		NewStep(1, id.StartTable(), NewIngressRelation(NewRelation("", nil, nil), false, false, "", ""), startRelationsList, startTableList, startCycles, 0),
	}
	log.Debug().Msg(fmt.Sprintf("%v", steps[0]))

	err = g.visitComponents(id.StartTable().Name(), func(r IngressRelation, comingFrom, goingTo Table, fromComponent, toComponent TableList, fromIndex, toIndex int, thisStepNumber, fromStepNumber uint) bool {
		subgraph := g.subGraph(toComponent)
		step := NewStep(thisStepNumber+1, goingTo, r, subgraph.relations, toComponent, subgraph.relCycles(goingTo), fromStepNumber+1)
		log.Debug().Msg(fmt.Sprintf("%v", step))
		steps = append(steps, step)
		return true
	})

	if err != nil {
		log.Warn().Msg(err.Error())
	}

	return NewPullerPlan(steps, g.relations, g.tables), nil
}

// Export the puller plan.
func Export(storage Storage, exporter Exporter) *Error {
	ep, err := GetPullerPlan(storage)
	if err != nil {
		return err
	}

	err = exporter.Export(ep)
	if err != nil {
		return err
	}

	return nil
}

// GetActiveTables returns list of tables that are activated (part of the connected graph)
func GetActiveTables(id IngressDescriptor) (TableList, *Error) {
	g := newGraph(id.Relations())
	return g.findConnectedTables(id.StartTable().Name())
}
