package pull

import (
	"fmt"

	"makeit.imfr.cgi.com/lino/pkg/pull"

	// import postgresql connector
	_ "github.com/lib/pq"
)

// PostgresDataSourceFactory exposes methods to create new Postgres pullers.
type PostgresDataSourceFactory struct {
	logger pull.Logger
}

// NewPostgresDataSourceFactory creates a new postgres datasource factory.
func NewPostgresDataSourceFactory(l pull.Logger) *PostgresDataSourceFactory {
	return &PostgresDataSourceFactory{l}
}

// New return a Postgres puller
func (e *PostgresDataSourceFactory) New(url string) pull.DataSource {
	return &SQLDataSource{
		url:     url,
		dialect: PostgresDialect{},
		logger:  e.logger,
	}
}

// PostgresDialect implement postgres SQL variations
type PostgresDialect struct{}

func (pd PostgresDialect) Placeholder(position int) string {
	return fmt.Sprintf("$%d", position)
}

func (pd PostgresDialect) Limit(limit uint) string {
	return fmt.Sprintf(" LIMIT %d", limit)
}
