package relation

import (
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/relation"

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
func (e *OracleExtractorFactory) New(url string) relation.Extractor {
	return NewOracleExtractor(url)
}

// OracleExtractor provides relation extraction logic from Oracle database.
type OracleExtractor struct {
	url string
}

// NewOracleExtractor creates a new oracle extractor.
func NewOracleExtractor(url string) *OracleExtractor {
	return &OracleExtractor{
		url: url,
	}
}

// Extract relations from the database.
func (e *OracleExtractor) Extract() ([]relation.Relation, *relation.Error) {
	db, err := dburl.Open(e.url)
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	SQL := `
	SELECT 
	a.owner schema, a.constraint_name name,
	c_pk.owner schema_source, c_pk.table_name parent_table, a_pk.COLUMN_NAME parent_key,
	a_pk.owner schema_source, a.table_name child_table, a.COLUMN_NAME child_key
FROM all_cons_columns a
JOIN all_constraints c ON a.owner = c.owner
					  AND a.constraint_name = c.constraint_name
JOIN all_constraints c_pk ON c.r_owner = c_pk.owner
						 AND c.r_constraint_name = c_pk.constraint_name
JOIN all_cons_columns a_pk ON c_pk.CONSTRAINT_NAME = a_pk.CONSTRAINT_NAME
						  AND a.POSITION = a_pk.POSITION
WHERE
a.owner =  user
ORDER by 1, 2 asc
            `

	rows, err := db.Query(SQL)
	if err != nil {
		return nil, &relation.Error{Description: err.Error()}
	}

	relations := []relation.Relation{}

	var (
		relationSchema string
		relationName   string
		sourceSchema   string
		sourceTable    string
		sourceColumn   string
		targetSchema   string
		targetTable    string
		targetColumn   string
	)

	for rows.Next() {
		err := rows.Scan(&relationSchema, &relationName, &sourceSchema, &sourceTable, &sourceColumn, &targetSchema, &targetTable, &targetColumn)
		if err != nil {
			return nil, &relation.Error{Description: err.Error()}
		}
		relation := relation.Relation{
			Name: relationName,
			Parent: relation.Table{
				Name: targetSchema + "." + targetTable,
				Keys: []string{targetColumn},
			},
			Child: relation.Table{
				Name: sourceSchema + "." + sourceTable,
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
