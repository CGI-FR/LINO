package push

// Table from which to push data.
type Table interface {
	Name() string
	PrimaryKey() string
}

// Plan describe how to push data
type Plan interface {
	FirstTable() Table
	RelationsFromTable(table Table) map[string]Relation
	Tables() []Table
}

// Relation between two tables.
type Relation interface {
	Name() string
	Parent() Table
	Child() Table
	OppositeOf(table Table) Table
}

// Value is an untyped data.
type Value interface{}

// Row of data.
type Row map[string]Value

// Error is the error type returned by the domain
type Error struct {
	Description string
}

func (e *Error) Error() string {
	return e.Description
}

// StopIteratorError signal the end of iterator
type StopIteratorError struct{}
