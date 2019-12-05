package table

// Table holds a name (table name) and a list of keys (table columns).
type Table struct {
	Name string
	Keys []string
}

// Error is the error type returned by the domain
type Error struct {
	Description string
}

func (e *Error) Error() string {
	return e.Description
}
