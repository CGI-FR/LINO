package id

func (g graph) findConnectedTables(start string) (TableList, *Error) {
	tables := []Table{}
	if err := g.visit(start, func(t Table) {
		tables = append(tables, t)
	}); err != nil {
		return nil, err
	}
	return NewTableList(tables), nil
}

func (g graph) getConnectedGraph(start string) (graph, *Error) {
	tables, err := g.findConnectedTables(start)
	if err != nil {
		return graph{}, err
	}
	return g.subGraph(tables), nil
}
