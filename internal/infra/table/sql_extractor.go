package table

import (
	"strings"

	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/table"
)

// SQLExtractor provides table extraction logic from SQL database.
type SQLExtractor struct {
	url     string
	schema  string
	dialect Dialect
}

type Dialect interface {
	SQL(schema string) string
}

// NewSQLExtractor creates a new SQL extractor.
func NewSQLExtractor(url string, schema string, dialect Dialect) *SQLExtractor {
	return &SQLExtractor{
		url:     url,
		schema:  schema,
		dialect: dialect,
	}
}

// Extract tables from the database.
func (e *SQLExtractor) Extract() ([]table.Table, *table.Error) {
	db, err := dburl.Open(e.url)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	SQL := e.dialect.SQL(e.schema)

	rows, err := db.Query(SQL)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

	tables := []table.Table{}

	var (
		tableSchema string
		tableName   string
		keyColumns  string
	)

	for rows.Next() {
		err := rows.Scan(&tableSchema, &tableName, &keyColumns)
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
