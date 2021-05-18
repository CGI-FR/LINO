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

// cycles applies the cycles detection algorithm
type cycles struct {
	graph      graph
	index      []Table
	blockedSet set
	blockedMap map[string][]Table
	stack      stack
	allCycles  [][]Table
}

// relCycles return all elementary relation cycles of the component
func (g graph) relCycles(startTable Table) CycleList {
	result := [][]IngressRelation{}
	tableCycles := g.cycles()
	log.Trace().Msg(fmt.Sprintf("cycle(s) %v", tableCycles))
	for _, tableCycle := range tableCycles {
		var idx int
		for i, table := range tableCycle {
			if table == startTable {
				idx = i
			}
		}
		tableCycle = append(tableCycle[idx:], tableCycle[:idx]...)
		tableCycle = append(tableCycle, startTable)
		log.Trace().Msg(fmt.Sprintf("rotated cycle is %v", tableCycle))
		result = append(result, g.develop(tableCycle)...)
	}
	list := []IngressRelationList{}
	for _, cycle := range result {
		list = append(list, NewIngressRelationList(cycle))
	}
	return NewCycleList(list)
}

// develop a table cycle to relation cycles
func (g graph) develop(cycle []Table) [][]IngressRelation {
	results := [][]IngressRelation{}
	if len(cycle) > 1 {
		from := cycle[0]
		to := cycle[1]
		relations := g.relationsFromTo(from, to)
		cycles := g.develop(cycle[1:])
		for _, rel := range relations {
			result := []IngressRelation{rel}
			for _, cycle := range cycles {
				result = append(result, cycle...)
			}
			set := newSet()
			resultDedup := []IngressRelation{}
			for _, rel := range result {
				if !set.contains(rel.Name()) {
					set.add(rel.Name())
					resultDedup = append(resultDedup, rel)
				}
			}
			results = append(results, resultDedup)
		}
	}
	return results
}

// cycles return all elementary cycles of the component
func (g graph) cycles() [][]Table {
	c := &cycles{
		graph: g,
		index: []Table{},
	}
	for _, v := range g.tabmap {
		c.index = append(c.index, v)
	}
	c.johnson()
	return c.allCycles
}

// johnson return all elementary cycles of the component using Johnson's algorithm
func (c *cycles) johnson() {
	var startIndex int
	for startIndex < len(c.graph.tabmap) {
		subGraph := c.subGraph(startIndex)
		components := subGraph.condense()
		leastIndexedTable, leastIndex, component := c.leastIndexedTableInMultiComponent(components)
		if leastIndexedTable == nil {
			break
		} else {
			// init and clear context (may contains data from previous loop)
			c.blockedSet = newSet()
			c.blockedMap = map[string][]Table{}
			c.stack = newStack()
			componentGraph := c.graph.subGraph(component)
			c.findCyclesInComponent(componentGraph, leastIndexedTable, leastIndexedTable)
			startIndex = leastIndex + 1
		}
	}
}

// subgraph create a subGraph ignoring table with index lower than startIndex
func (c *cycles) subGraph(startIndex int) graph {
	if startIndex == 0 {
		return c.graph
	}

	subGraph := []IngressRelation{}
	set := newSet()

	for idx, v := range c.index {
		if idx >= startIndex {
			set.add(v.Name())
		}
	}
	for i := uint(0); i < c.graph.relations.Len(); i++ {
		rel := c.graph.relations.Relation(i)
		parentIn := set.contains(rel.Parent().Name())
		childIn := set.contains(rel.Child().Name())
		if parentIn && childIn {
			subGraph = append(subGraph, rel)
		}
	}

	return newGraph(NewIngressRelationList(subGraph))
}

// leastIndexedTableInMultiComponent returns the first table part of a multi-table component and its index (or nil)
func (c *cycles) leastIndexedTableInMultiComponent(components []TableList) (Table, int, TableList) {
	multiComponents := []TableList{}
	for _, component := range components {
		if component.Len() > 1 {
			multiComponents = append(multiComponents, component)
		}
	}
	for idx, v := range c.index {
		// in which component is table
		for _, component := range multiComponents {
			if ok := component.Contains(v.Name()); ok {
				return v, idx, component
			}
		}
	}
	return nil, 0, nil
}

func (c *cycles) findCyclesInComponent(search graph, startTable Table, currentTable Table) bool {
	foundCycle := false
	c.stack.push(currentTable)
	c.blockedSet.add(currentTable.Name())
	for _, relname := range search.activefrom[currentTable.Name()] {
		relation := search.relmap[relname]
		var childTable Table
		if currentTable.Name() == relation.Child().Name() {
			childTable = relation.Parent()
		} else {
			childTable = relation.Child()
		}
		if childTable.Name() == startTable.Name() {
			// cycle is found, store it
			c.allCycles = append(c.allCycles, c.stack.asSlice())
			foundCycle = true
		} else if !c.blockedSet.contains(childTable.Name()) {
			// explore this child table only if not in blockedSet
			gotCycle := c.findCyclesInComponent(search, startTable, childTable)
			foundCycle = foundCycle || gotCycle
		}
	}
	if foundCycle {
		c.unblock(currentTable)
	} else {
		for _, relname := range search.activefrom[currentTable.Name()] {
			relation := search.relmap[relname]
			var childTable Table
			if currentTable.Name() == relation.Child().Name() {
				childTable = relation.Parent()
			} else {
				childTable = relation.Child()
			}
			c.blockedMap[childTable.Name()] = append(c.blockedMap[childTable.Name()], currentTable)
		}
	}
	c.stack.pop()
	return foundCycle
}

func (c *cycles) unblock(t Table) {
	c.blockedSet.remove(t.Name())
	if tableList, ok := c.blockedMap[t.Name()]; ok {
		for _, toUnblock := range tableList {
			c.unblock(toUnblock)
		}
		delete(c.blockedMap, t.Name())
	}
}
