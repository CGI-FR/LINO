package push

import (
	"makeit.imfr.cgi.com/lino/pkg/push"
	// import postgresql connector
)

// PostgresDataDestinationFactory exposes methods to create new Postgres pullers.
type PostgresDataDestinationFactory struct {
	logger push.Logger
}

// NewPostgresDataDestinationFactory creates a new postgres datadestination factory.
func NewPostgresDataDestinationFactory(l push.Logger) *PostgresDataDestinationFactory {
	return &PostgresDataDestinationFactory{l}
}

// New return a Postgres pusher
func (e *PostgresDataDestinationFactory) New(url string) push.DataDestination {
	return NewSQLDataDestination(url, e.logger)
}
