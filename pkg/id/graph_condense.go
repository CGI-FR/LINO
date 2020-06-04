package id

// condenser applies the condensation algorithm
type condenser struct {
	graph      graph
	stack      stack
	index      uint
	components []TableList
	vindex     map[string]uint
	vlowlink   map[string]uint
	vonstack   map[string]bool
}

// condense the graph
func (g graph) condense() []TableList {
	c := &condenser{
		graph:      g,
		stack:      newStack(),
		components: []TableList{},
		vindex:     map[string]uint{},
		vlowlink:   map[string]uint{},
		vonstack:   map[string]bool{},
	}

	c.tarjan()

	return c.components
}

func (c *condenser) tarjan() {
	for _, v := range c.graph.tabmap {
		_, ok := c.vindex[v.Name()]
		if !ok {
			c.strongconnect(v)
		}
	}
}

func (c *condenser) strongconnect(v Table) {
	// Set the depth index for v to the smallest unused index
	c.vindex[v.Name()] = c.index
	c.vlowlink[v.Name()] = c.index
	c.index++
	c.stack.push(v)
	c.vonstack[v.Name()] = true

	// Consider successors of v
	for _, relname := range c.graph.activefrom[v.Name()] {
		e := c.graph.relmap[relname]
		var w Table
		if v.Name() != e.Child().Name() {
			w = e.Child()
		} else {
			w = e.Parent()
		}

		_, ok := c.vindex[w.Name()]
		if !ok {
			// Successor w has not yet been visited; recurse on it
			c.strongconnect(w)
			c.vlowlink[v.Name()] = umin(c.vlowlink[v.Name()], c.vlowlink[w.Name()])
		} else if c.vonstack[w.Name()] {
			// Successor w is in stack S and hence in the current SCC
			// If w is not on stack, then (v, w) is a cross-edge in the DFS tree and must be ignored
			// Note: The next line may look odd - but is correct.
			// It says w.index not w.lowlink; that is deliberate and from the original paper
			c.vlowlink[v.Name()] = umin(c.vlowlink[v.Name()], c.vindex[w.Name()])
		}
	}

	// If v is a root node, pop the stack and generate an SCC
	if c.vlowlink[v.Name()] == c.vindex[v.Name()] {
		// start a new strongly connected component
		tables := []Table{}
		for {
			w := c.stack.pop()
			c.vonstack[w.Name()] = false
			//add w to current strongly connected component
			tables = append(tables, w)
			if w.Name() == v.Name() {
				break
			}
		}
		c.components = append(c.components, NewTableList(tables))
	}
}

func umin(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}
