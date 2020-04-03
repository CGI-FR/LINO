package push

type relation struct {
	name   string
	parent Table
	child  Table
}

// NewRelation initialize a new Relation object
func NewRelation(name string, parent Table, child Table) Relation {
	return relation{name: name, parent: parent, child: child}
}

func (r relation) Name() string   { return r.name }
func (r relation) Parent() Table  { return r.parent }
func (r relation) Child() Table   { return r.child }
func (r relation) String() string { return r.name }
func (r relation) OppositeOf(table Table) Table {
	if r.Child().Name() == table.Name() {
		return r.Parent()
	}
	return r.Child()
}
