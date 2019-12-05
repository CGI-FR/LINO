package dataconnector

// DataConnector holds a name (alias) and a URI to a database.
type DataConnector struct {
	Name string
	URL  string
}

// Error is the error type returned by the domain
type Error struct {
	Description string
}

func (e *Error) Error() string {
	return e.Description
}
