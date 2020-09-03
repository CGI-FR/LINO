package pull

import (
	"fmt"

	"makeit.imfr.cgi.com/lino/pkg/pull"

	// import oracle connector
	_ "github.com/lib/pq"
)

// OracleDataSourceFactory exposes methods to create new Oracle pullers.
type OracleDataSourceFactory struct {
	logger pull.Logger
}

// NewOracleDataSourceFactory creates a new oracle datasource factory.
func NewOracleDataSourceFactory(l pull.Logger) *OracleDataSourceFactory {
	return &OracleDataSourceFactory{l}
}

// New return a Oracle puller
func (e *OracleDataSourceFactory) New(url string, schema string) pull.DataSource {
	return &SQLDataSource{
		url:     url,
		schema:  schema,
		dialect: PostgresDialect{},
		logger:  e.logger,
	}
}

// PostgresDialect implement postgres SQL variations
type OracleDialect struct{}

func (od OracleDialect) Placeholder(position int) string {
	return fmt.Sprintf(":v%d", position)
}

func (od OracleDialect) Limit(limit uint) string {
	return fmt.Sprintf("AND rownum <= %v", limit)
}
