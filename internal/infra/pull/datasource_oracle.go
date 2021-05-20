package pull

import (
	"fmt"

	"github.com/cgi-fr/lino/pkg/pull"
)

// OracleDataSourceFactory exposes methods to create new Oracle pullers.
type OracleDataSourceFactory struct{}

// NewOracleDataSourceFactory creates a new oracle datasource factory.
func NewOracleDataSourceFactory() *OracleDataSourceFactory {
	return &OracleDataSourceFactory{}
}

// New return a Oracle puller
func (e *OracleDataSourceFactory) New(url string, schema string) pull.DataSource {
	return &SQLDataSource{
		url:     url,
		schema:  schema,
		dialect: OracleDialect{},
	}
}

// PostgresDialect implement postgres SQL variations
type OracleDialect struct{}

func (od OracleDialect) Placeholder(position int) string {
	return fmt.Sprintf(":v%d", position)
}

func (od OracleDialect) Limit(limit uint) string {
	return fmt.Sprintf(" AND rownum <= %d", limit)
}
