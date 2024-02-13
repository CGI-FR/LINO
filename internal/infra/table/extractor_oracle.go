package table

import (
	"fmt"

	// import Oracle connector
	_ "github.com/sijms/go-ora/v2"

	"github.com/cgi-fr/lino/pkg/table"
)

// NewOracleExtractorFactory creates a new oracle extractor factory.
func NewOracleExtractorFactory() *OracleExtractorFactory {
	return &OracleExtractorFactory{}
}

// OracleExtractorFactory exposes methods to create new Oracle extractors.
type OracleExtractorFactory struct{}

// New return a Oracle extractor
func (e *OracleExtractorFactory) New(url string, schema string) table.Extractor {
	return NewSQLExtractor(url, schema, OracleDialect{})
}

type OracleDialect struct{}

func (d OracleDialect) SQL(schema string) string {
	SQL := `
SELECT
	all_cons_columns.owner as schema_name,
	all_cons_columns.table_name as table_name,
	LISTAGG(all_cons_columns.column_name, ',') WITHIN GROUP (order by all_cons_columns.position) as columns
 FROM all_constraints, all_cons_columns
 where
	all_constraints.constraint_type = 'P'
	and all_constraints.constraint_name = all_cons_columns.constraint_name
	and all_constraints.owner = all_cons_columns.owner
	`

	if schema == "" {
		SQL += "AND all_constraints.owner =  user"
	} else {
		SQL += fmt.Sprintf("AND all_constraints.owner = '%s'", schema)
	}

	SQL += `
 group by all_cons_columns.table_name, all_cons_columns.owner
	`

	return SQL
}

func (d OracleDialect) Select(tableName string) string {
	query := "SELECT * FROM "
	query += tableName
	query += " WHERE ROWNUM <= 0"

	return query
}
