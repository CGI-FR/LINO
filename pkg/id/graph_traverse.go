package id

type traversal struct {
	g       graph
	visit   func(Table)
	visited set
}

// visit each table starting from the table named start.
func (g graph) visit(start string, visit func(Table)) *Error {
	t := &traversal{g, visit, newSet()}
	startTable, ok := t.g.tabmap[start]
	if !ok {
		return &Error{Description: "no table named " + start}
	}
	return t.follow(startTable)
}

// visit each table starting from the table named start following children tables only.
func (g graph) visitChildren(start string, visit func(Table)) *Error {
	t := &traversal{g, visit, newSet()}
	startTable, ok := t.g.tabmap[start]
	if !ok {
		return &Error{Description: "no table named " + start}
	}
	return t.followChilren(startTable)
}

// visit each table starting from the table named start following active relations only.
/* func (g graph) visitActive(start string, visit func(r IngressRelation, comingFrom, goingTo Table, fromComponent, toComponent TableList, fromIndex, toIndex int, crossingBorder bool, fromStep, thisStep uint) bool) *Error {
	t := &reltraversal{g, g.condense(), visit, map[int]set{}, 0}
	startTable, ok := t.g.tabmap[start]
	if !ok {
		return &Error{Description: "no table named " + start}
	}
	return t.followActive(startTable, 0)
} */

// visit each component starting from the table named start following active relations only.
func (g graph) visitComponents(start string, visit func(r IngressRelation, comingFrom, goingTo Table, fromComponent, toComponent TableList, fromIndex, toIndex int, thisStepNumber, fromStepNumber uint) bool) *Error {
	renumber := map[uint]uint{}
	number := uint(0)
	myvisit := func(r IngressRelation, comingFrom, goingTo Table, fromComponent, toComponent TableList, fromIndex, toIndex int, crossingBorder bool, fromStep, thisStep uint) bool {
		if crossingBorder {
			number++
		}

		// renumber steps
		newNumber, ok := renumber[thisStep]
		if !ok {
			newNumber = number
			renumber[thisStep] = newNumber
		}
		thisStep = newNumber

		if crossingBorder {
			return visit(r, comingFrom, goingTo, fromComponent, toComponent, fromIndex, toIndex, thisStep, renumber[fromStep])
		}
		return true
	}
	t := &reltraversal{g, g.condense(), myvisit, map[int]set{}, 0}
	startTable, ok := t.g.tabmap[start]
	if !ok {
		return &Error{Description: "table named " + start + " not connectected with relation(s) " + t.g.relations.String()}
	}
	return t.followActive(startTable, 0)
}

func (t *traversal) follow(from Table) *Error {
	t.visited.add(from.Name())
	t.visit(from)

	for _, next := range t.g.neighboursOf(from) {
		if !t.visited.contains(next.Name()) {
			if err := t.follow(next); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *traversal) followChilren(from Table) *Error {
	t.visited.add(from.Name())
	t.visit(from)

	for _, next := range t.g.childrenOf(from) {
		if !t.visited.contains(next.Name()) {
			if err := t.followChilren(next); err != nil {
				return err
			}
		}
	}

	return nil
}

type reltraversal struct {
	g          graph
	components []TableList
	visit      func(r IngressRelation, comingFrom, goingTo Table, fromComponent, toComponent TableList, fromIndex, toIndex int, crossingBorder bool, fromStep, thisStep uint) bool
	visitedmap map[int]set
	stepNumber uint
}

func (t *reltraversal) followActive(from Table, fromStep uint) *Error {
	relations := t.g.activefrom[from.Name()]
	relin := []string{}  // inner relationships don't cross component border, we process them first
	relout := []string{} // outer relationships cross component border, we process them after inner relationships

	for _, relname := range relations {
		rel := t.g.relmap[relname]
		next := rel.Child()
		if next == from {
			next = rel.Parent()
		}
		indexFrom, _ := findEnclosingComponent(from, t.components)
		indexTo, _ := findEnclosingComponent(next, t.components)
		if indexFrom == indexTo {
			relin = append(relin, relname)
		} else {
			relout = append(relout, relname)
		}
	}

	for _, relname := range relin {
		t.stepNumber++
		thisStep := t.stepNumber
		rel := t.g.relmap[relname]
		next := rel.Child()
		if next == from {
			next = rel.Parent()
		}
		indexFrom, componentFrom := findEnclosingComponent(from, t.components)
		indexTo, componentTo := findEnclosingComponent(next, t.components)
		if indexFrom != indexTo {
			panic("coding error")
		}
		visited, ok := t.visitedmap[indexFrom]
		if !ok {
			visited = newSet()
		}
		t.visitedmap[indexFrom] = visited
		visited.add(from.Name())
		if t.visit(rel, from, next, componentFrom, componentTo, indexFrom, indexTo, indexFrom != indexTo, fromStep, thisStep) && !visited.contains(next.Name()) {
			if err := t.followActive(next, thisStep); err != nil {
				return err
			}
		}
	}
	for _, relname := range relout {
		t.stepNumber++
		thisStep := t.stepNumber
		rel := t.g.relmap[relname]
		next := rel.Child()
		if next == from {
			next = rel.Parent()
		}
		indexFrom, componentFrom := findEnclosingComponent(from, t.components)
		indexTo, componentTo := findEnclosingComponent(next, t.components)
		if t.visit(rel, from, next, componentFrom, componentTo, indexFrom, indexTo, indexFrom != indexTo, fromStep, thisStep) {
			if err := t.followActive(next, thisStep); err != nil {
				return err
			}
		}
	}

	return nil
}

func findEnclosingComponent(t Table, components []TableList) (int, TableList) {
	for index, component := range components {
		if component.Contains(t.Name()) {
			return index, component
		}
	}
	return -1, nil
}
