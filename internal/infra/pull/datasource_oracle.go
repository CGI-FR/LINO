package pull

import (
	"github.com/cgi-fr/lino/internal/infra/commonsql"
	"github.com/cgi-fr/lino/pkg/pull"
)

// OracleDataSourceFactory exposes methods to create new Oracle pullers.
type OracleDataSourceFactory struct{}

// NewOracleDataSourceFactory creates a new oracle datasource factory.
func NewOracleDataSourceFactory() *OracleDataSourceFactory {
	return &OracleDataSourceFactory{}
}

// New return a Oracle puller
func (e *OracleDataSourceFactory) New(url string, schema string, options ...pull.DataSourceOption) pull.DataSource {
	ds := &SQLDataSource{
		url:     url,
		schema:  schema,
		dialect: commonsql.OracleDialect{},
	}

	for _, option := range options {
		option(ds)
	}

	return ds
}
