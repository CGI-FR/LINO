package load

type plan struct {
	firstTable Table
	relations  []Relation
}

// NewPlan initialize a new Plan object
func NewPlan(first Table, relations []Relation) Plan {
	return plan{firstTable: first, relations: relations}
}

func (p plan) FirstTable() Table { return p.firstTable }
func (p plan) RelationsFromTable(table Table) map[string]Relation {
	result := map[string]Relation{}
	for _, r := range p.relations {
		if r.Parent().Name() == table.Name() {
			result[r.Name()] = r
		}
	}
	return result
}
func (p plan) Tables() []Table {
	result := []Table{}

	tables := map[string]Table{}
	for _, r := range p.relations {
		tables[r.Child().Name()] = r.Child()
		tables[r.Parent().Name()] = r.Parent()
	}

	for _, v := range tables {
		result = append(result, v)
	}
	return result
}
