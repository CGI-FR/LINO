package relation

import (

	// import postgresql connector
	"fmt"

	_ "github.com/lib/pq"

	"makeit.imfr.cgi.com/lino/pkg/relation"
)

// NewPostgresExtractorFactory creates a new postgres extractor factory.
func NewPostgresExtractorFactory() *PostgresExtractorFactory {
	return &PostgresExtractorFactory{}
}

// PostgresExtractorFactory exposes methods to create new Postgres extractors.
type PostgresExtractorFactory struct{}

// New return a Postgres extractor
func (e *PostgresExtractorFactory) New(url string, schema string) relation.Extractor {
	return NewSQLExtractor(url, schema, PostgresDialect{})
}

type PostgresDialect struct{}

func (d PostgresDialect) SQL(schema string) string {
	SQL := `
SELECT
    tc.constraint_name,
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM
    information_schema.table_constraints AS tc
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
      AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
            `

	if schema != "" {
		SQL += fmt.Sprintf("AND tc.table_schema = '%s'", schema)
	}
	return SQL
}
