package relation

import (
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/relation"
)

// SQLExtractor provides relation extraction logic from SQL database.
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

// Extract relations from the database.
func (e *SQLExtractor) Extract() ([]relation.Relation, *relation.Error) {
	db, err := dburl.Open(e.url)
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	rows, err := db.Query(e.dialect.SQL(e.schema))
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	relations := []relation.Relation{}

	var (
		relationName string
		sourceTable  string
		sourceColumn string
		targetTable  string
		targetColumn string
	)

	for rows.Next() {
		err := rows.Scan(&relationName, &sourceTable, &sourceColumn, &targetTable, &targetColumn)
		if err != nil {
			return nil, &relation.Error{Description: err.Error()}
		}

		relation := relation.Relation{
			Name: relationName,
			Parent: relation.Table{
				Name: targetTable,
				Keys: []string{targetColumn},
			},
			Child: relation.Table{
				Name: sourceTable,
				Keys: []string{sourceColumn},
			},
		}
		relations = append(relations, relation)
	}
	err = rows.Err()
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	return relations, nil
}
