package pull

import (
	"fmt"

	// import db2 connector
	"github.com/cgi-fr/lino/pkg/pull"
	_ "github.com/ibmdb/go_ibm_db"
)

// Db2DataSourceFactory exposes methods to create new Db2 pullers.
type Db2DataSourceFactory struct{}

// NewDb2DataSourceFactory creates a new oracle datasource factory.
func NewDb2DataSourceFactory() *Db2DataSourceFactory {
	return &Db2DataSourceFactory{}
}

// New return a Db2 puller
func (e *Db2DataSourceFactory) New(url string, schema string) pull.DataSource {
	return &SQLDataSource{
		url:     url,
		schema:  schema,
		dialect: Db2Dialect{},
	}
}

// PostgresDialect implement postgres SQL variations
type Db2Dialect struct{}

func (od Db2Dialect) Placeholder(position int) string {
	return "?"
}

func (od Db2Dialect) Limit(limit uint) string {
	return fmt.Sprintf(" LIMIT %d", limit)
}
