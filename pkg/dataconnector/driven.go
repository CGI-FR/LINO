package dataconnector

// Storage allows to store and retrieve DataConnector objects.
type Storage interface {
	List() ([]DataConnector, *Error)
	Store(*DataConnector) *Error
}
