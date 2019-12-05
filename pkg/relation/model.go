package relation

// Table holds a name (table name) and a list of keys (table columns).
type Table struct {
	Name string
	Keys []string
}

// Relation holds a parent Table and a child Table
type Relation struct {
	Name   string
	Parent Table
	Child  Table
}

// Error is the error type returned by the domain
type Error struct {
	Description string
}

func (e *Error) Error() string {
	return e.Description
}
