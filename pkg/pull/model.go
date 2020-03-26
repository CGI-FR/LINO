package pull

// Table from which to pull data.
type Table interface {
	Name() string
	PrimaryKey() string
}

// Relation between two tables.
type Relation interface {
	Name() string
	Parent() Table
	Child() Table
	ParentKey() string
	ChildKey() string
	OppositeOf(tablename string) Table
}

// RelationList is a list of relations.
type RelationList interface {
	Len() uint
	Relation(idx uint) Relation
}

// Cycle is a list of relations.
type Cycle interface {
	RelationList
}

// CycleList is a list of cycles.
type CycleList interface {
	Len() uint
	Cycle(idx uint) Cycle
}

// Step group of follows to perform.
type Step interface {
	Index() uint
	Entry() Table
	Follow() Relation
	Relations() RelationList
	Cycles() CycleList
	NextSteps() StepList
}

// StepList list of steps to perform.
type StepList interface {
	Len() uint
	Step(uint) Step
}

// Plan of the pullion process.
type Plan interface {
	InitFilter() Filter
	Steps() StepList
}

// Value is an untyped data.
type Value interface{}

// Filter applied to data tables.
type Filter interface {
	Limit() uint
	Values() Row
}

// Row of data.
type Row map[string]Value

// Error is the error type returned by the domain
type Error struct {
	Description string
}

func (e *Error) Error() string {
	return e.Description
}
