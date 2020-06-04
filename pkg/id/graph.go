package id

type graph struct {
	relations IngressRelationList
	tables    TableList

	// tables and relations indexed by name
	tabmap map[string]Table
	relmap map[string]IngressRelation

	// active relations, key=tablename, value=set of relation names
	activefrom map[string][]string // relations going out of the table
	activeto   map[string][]string // relations going in the table

	// tables, key=tableblename, value=set of table names
	neighbours map[string]set
	parents    map[string]set
	children   map[string]set
}

func newGraph(rels IngressRelationList) graph {
	g := graph{
		relations:  rels,
		tables:     nil,
		tabmap:     map[string]Table{},
		relmap:     map[string]IngressRelation{},
		activefrom: map[string][]string{},
		activeto:   map[string][]string{},
		neighbours: map[string]set{},
		parents:    map[string]set{},
		children:   map[string]set{},
	}

	for i := uint(0); i < rels.Len(); i++ {
		r := rels.Relation(i)
		g.addRelation(r)
	}

	tables := []Table{}
	for _, table := range g.tabmap {
		tables = append(tables, table)
	}
	g.tables = NewTableList(tables)

	return g
}

/* func (g graph) addTable(t Table) {
	g.tabmap[t.Name()] = t
} */

/* func (g graph) containsTable(t Table) bool {
	_, ok := g.tabmap[t.Name()]
	return ok
} */

func (g graph) Len() uint {
	return g.relations.Len()
}

func (g graph) addRelation(r IngressRelation) {
	parentName := r.Parent().Name()
	childName := r.Child().Name()

	g.tabmap[parentName] = r.Parent()
	g.tabmap[childName] = r.Child()
	g.relmap[r.Name()] = r

	g.link(parentName, childName)
	g.link(childName, parentName)

	parents, ok := g.parents[childName]
	if !ok {
		parents = newSet()
	}
	parents.add(parentName)
	g.parents[childName] = parents

	children, ok := g.children[parentName]
	if !ok {
		children = newSet()
	}
	children.add(childName)
	g.children[parentName] = children

	if r.LookUpChild() {
		activefrom, ok := g.activefrom[parentName]
		if !ok {
			activefrom = []string{}
		}
		activefrom = append(activefrom, r.Name())
		g.activefrom[parentName] = activefrom

		activeto, ok := g.activeto[childName]
		if !ok {
			activeto = []string{}
		}
		activeto = append(activeto, r.Name())
		g.activeto[childName] = activeto
	}

	if r.LookUpParent() {
		activefrom, ok := g.activefrom[childName]
		if !ok {
			activefrom = []string{}
		}
		activefrom = append(activefrom, r.Name())
		g.activefrom[childName] = activefrom

		activeto, ok := g.activeto[parentName]
		if !ok {
			activeto = []string{}
		}
		activeto = append(activeto, r.Name())
		g.activeto[parentName] = activeto
	}
}

func (g graph) link(from, to string) {
	neighbours, ok := g.neighbours[from]
	if !ok {
		neighbours = newSet()
	}
	neighbours.add(to)
	g.neighbours[from] = neighbours
}

func (g graph) neighboursOf(t Table) []Table {
	result := []Table{}
	for name := range g.neighbours[t.Name()] {
		result = append(result, g.tabmap[name])
	}
	return result
}

func (g graph) childrenOf(t Table) []Table {
	result := []Table{}
	for name := range g.children[t.Name()] {
		result = append(result, g.tabmap[name])
	}
	return result
}

/* func (g graph) parentsOf(t Table) []Table {
	result := []Table{}
	for name := range g.parents[t.Name()] {
		result = append(result, g.tabmap[name])
	}
	return result
} */

func (g graph) relationsFrom(t Table) []Relation {
	result := []Relation{}
	for i := uint(0); i < g.relations.Len(); i++ {
		rel := g.relations.Relation(i)
		if rel.Parent().Name() == t.Name() {
			result = append(result, rel)
		}
	}
	return result
}

/* func (g graph) relationsTo(t Table) []Relation {
	result := []Relation{}
	for _, rel := range g.order {
		if rel.Child().Name() == t.Name() {
			result = append(result, rel)
		}
	}
	return result
} */

func (g graph) relationsFromTo(from, to Table) []IngressRelation {
	result := []IngressRelation{}
	for i := uint(0); i < g.relations.Len(); i++ {
		rel := g.relations.Relation(i)
		if rel.Parent().Name() == from.Name() && rel.Child().Name() == to.Name() && rel.LookUpChild() {
			result = append(result, rel)
		}
		if rel.Child().Name() == from.Name() && rel.Parent().Name() == to.Name() && rel.LookUpParent() {
			result = append(result, rel)
		}
	}
	return result
}

func (g graph) subGraph(tables TableList) graph {
	result := []IngressRelation{}
	for i := uint(0); i < g.relations.Len(); i++ {
		rel := g.relations.Relation(i)
		if tables.Contains(rel.Child().Name()) && tables.Contains(rel.Parent().Name()) {
			result = append(result, rel)
		}
	}
	return newGraph(NewIngressRelationList(result))
}

func (g graph) slim() graph {
	activeRelations := []IngressRelation{}
	for i := uint(0); i < g.relations.Len(); i++ {
		rel := g.relations.Relation(i)
		if rel.LookUpChild() || rel.LookUpParent() {
			activeRelations = append(activeRelations, rel)
		}
	}
	return newGraph(NewIngressRelationList(activeRelations))
}
