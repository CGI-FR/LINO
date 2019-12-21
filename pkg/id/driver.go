package id

import "fmt"

var logger Logger = Nologger{}

// SetLogger if needed, default no logger
func SetLogger(l Logger) {
	logger = l
}

// Create and store ingress descriptor for the given start table and relation set.
func Create(startTable string, relReader RelationReader, storage Storage) *Error {
	relations, err := relReader.Read()
	if err != nil {
		return err
	}

	ingressRels := []IngressRelation{}
	for i := uint(0); i < relations.Len(); i++ {
		rel := relations.Relation(i)
		ingressRels = append(ingressRels, NewIngressRelation(rel, false, false))
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
			adrelations = append(adrelations, NewIngressRelation(NewRelation(rel.Name(), rel.Parent(), rel.Child()), false, true))
		} else {
			adrelations = append(adrelations, NewIngressRelation(NewRelation(rel.Name(), rel.Parent(), rel.Child()), false, false))
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

// GetExtractionPlan returns the calculated extraction plan.
func GetExtractionPlan(storage Storage) (ExtractionPlan, *Error) {
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
		logger.Debug(fmt.Sprintf("component %v - %v - %v", i, component, cycles))
		if component.Contains(id.StartTable().Name()) {
			startComponent = component
		}
	}
	logger.Debug("")

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
		NewStep(1, id.StartTable(), NewIngressRelation(NewRelation("", nil, nil), false, false), startRelationsList, startTableList, startCycles, 0),
	}
	logger.Debug(fmt.Sprintf("%v", steps[0]))

	err = g.visitComponents(id.StartTable().Name(), func(r IngressRelation, comingFrom, goingTo Table, fromComponent, toComponent TableList, fromIndex, toIndex int, thisStepNumber, fromStepNumber uint) bool {
		subgraph := g.subGraph(toComponent)
		step := NewStep(thisStepNumber+1, goingTo, r, subgraph.relations, toComponent, subgraph.relCycles(goingTo), fromStepNumber+1)
		logger.Debug(fmt.Sprintf("%v", step))
		steps = append(steps, step)
		return true
	})
	logger.Debug("")

	if err != nil {
		return nil, err
	}

	return NewExtractionPlan(steps, g.relations, g.tables), nil
}

// Export the extraction plan.
func Export(storage Storage, exporter Exporter) *Error {
	ep, err := GetExtractionPlan(storage)
	if err != nil {
		return err
	}

	err = exporter.Export(ep)
	if err != nil {
		return err
	}

	return nil
}
