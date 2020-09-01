package table

import (
	"fmt"
	"strings"

	// import postgresql connector
	_ "github.com/lib/pq"

	"github.com/xo/dburl"
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
	return NewPostgresExtractor(url, schema)
}

// PostgresExtractor provides table extraction logic from Postgres database.
type PostgresExtractor struct {
	url    string
	schema string
}

// NewPostgresExtractor creates a new postgres extractor.
func NewPostgresExtractor(url string, schema string) *PostgresExtractor {
	return &PostgresExtractor{
		url:    url,
		schema: schema,
	}
}

// Extract tables from the database.
func (e *PostgresExtractor) Extract() ([]table.Table, *table.Error) {
	db, err := dburl.Open(e.url)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	SQL := `SELECT kcu.table_schema,
		kcu.table_name,
		tco.constraint_name,
		string_agg(kcu.column_name,', ') AS key_columns
	FROM information_schema.table_constraints tco
	JOIN information_schema.key_column_usage kcu
	ON kcu.constraint_name = tco.constraint_name
	AND kcu.constraint_schema = tco.constraint_schema
	AND kcu.constraint_name = tco.constraint_name
	WHERE tco.constraint_type = 'PRIMARY KEY'
	`

	if e.schema != "" {
		SQL += fmt.Sprintf("AND kcu.table_schema = '%s'", e.schema)
	}

	SQL += `
	GROUP BY tco.constraint_name,
		kcu.table_schema,
		kcu.table_name
	ORDER BY kcu.table_schema,
		kcu.table_name`

	rows, err := db.Query(SQL)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	tables := []table.Table{}

	var (
		tableSchema    string
		tableName      string
		constraintName string
		keyColumns     string
	)

	for rows.Next() {
		err := rows.Scan(&tableSchema, &tableName, &constraintName, &keyColumns)
		if err != nil {
			return nil, &table.Error{Description: err.Error()}
		}

		table := table.Table{

			Name: tableName,
			Keys: strings.Split(keyColumns, ", "),
		}
		tables = append(tables, table)
	}
	err = rows.Err()
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	return tables, nil
}
