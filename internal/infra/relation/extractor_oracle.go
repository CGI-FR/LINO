package relation

import (
	"fmt"

	"github.com/cgi-fr/lino/pkg/relation"

	// import Oracle connector
	_ "github.com/godror/godror"
)

// NewOracleExtractorFactory creates a new oracle extractor factory.
func NewOracleExtractorFactory() *OracleExtractorFactory {
	return &OracleExtractorFactory{}
}

// OracleExtractorFactory exposes methods to create new Oracle extractors.
type OracleExtractorFactory struct{}

// New return a Oracle extractor
func (e *OracleExtractorFactory) New(url string, schema string) relation.Extractor {
	return NewSQLExtractor(url, schema, OracleDialect{})
}

type OracleDialect struct{}

func (d OracleDialect) SQL(schema string) string {
	SQL := `
	SELECT
	a.constraint_name name,
	a.table_name child_table,
	a.COLUMN_NAME child_key,
	c_pk.table_name parent_table,
	a_pk.COLUMN_NAME parent_key
FROM all_cons_columns a
JOIN all_constraints c ON a.owner = c.owner
					  AND a.constraint_name = c.constraint_name
JOIN all_constraints c_pk ON c.r_owner = c_pk.owner
						 AND c.r_constraint_name = c_pk.constraint_name
JOIN all_cons_columns a_pk ON c_pk.CONSTRAINT_NAME = a_pk.CONSTRAINT_NAME
						  AND a.POSITION = a_pk.POSITION
WHERE
`

	if schema == "" {
		SQL += "a.owner = user"
	} else {
		SQL += fmt.Sprintf("a.owner = '%s'", schema)
	}

	SQL += `
ORDER by 1, 2 asc
`
	return SQL
}
