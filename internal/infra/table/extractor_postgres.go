package table

import (
	"fmt"

	// import postgresql connector
	_ "github.com/lib/pq"

	"makeit.imfr.cgi.com/lino/pkg/table"
)

// NewPostgresExtractorFactory creates a new postgres extractor factory.
func NewPostgresExtractorFactory() *PostgresExtractorFactory {
	return &PostgresExtractorFactory{}
}

// PostgresExtractorFactory exposes methods to create new Postgres extractors.
type PostgresExtractorFactory struct{}

// New return a Postgres extractor
func (e *PostgresExtractorFactory) New(url string, schema string) table.Extractor {
	return NewSQLExtractor(url, schema, PostgresDialect{})
}

type PostgresDialect struct {
}

func (d PostgresDialect) SQL(schema string) string {
	SQL := `SELECT kcu.table_schema,
	kcu.table_name,
	string_agg(kcu.column_name,', ') AS key_columns
FROM information_schema.table_constraints tco
JOIN information_schema.key_column_usage kcu
ON kcu.constraint_name = tco.constraint_name
AND kcu.constraint_schema = tco.constraint_schema
AND kcu.constraint_name = tco.constraint_name
WHERE tco.constraint_type = 'PRIMARY KEY'
`

	if schema != "" {
		SQL += fmt.Sprintf("AND kcu.table_schema = '%s'", schema)
	}

	SQL += `
GROUP BY tco.constraint_name,
	kcu.table_schema,
	kcu.table_name
ORDER BY kcu.table_schema,
	kcu.table_name`

	return SQL
}
