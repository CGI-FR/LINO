package table

import (
	"strings"

	// import Oracle connector
	_ "github.com/godror/godror"

	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/table"
)

// NewOracleExtractorFactory creates a new oracle extractor factory.
func NewOracleExtractorFactory() *OracleExtractorFactory {
	return &OracleExtractorFactory{}
}

// OracleExtractorFactory exposes methods to create new Oracle extractors.
type OracleExtractorFactory struct{}

// New return a Oracle extractor
func (e *OracleExtractorFactory) New(url string) table.Extractor {
	return NewOracleExtractor(url)
}

// OracleExtractor provides table extraction logic from Oracle database.
type OracleExtractor struct {
	url string
}

// NewOracleExtractor creates a new oracle extractor.
func NewOracleExtractor(url string) *OracleExtractor {
	return &OracleExtractor{
		url: url,
	}
}

// Extract tables from the database.
func (e *OracleExtractor) Extract() ([]table.Table, *table.Error) {
	db, err := dburl.Open(e.url)
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, &table.Error{Description: err.Error()}
	}

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
	and all_constraints.owner =  user
	
 group by all_cons_columns.table_name, all_cons_columns.owner
	`

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
			Name: tableSchema + "." + tableName,
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
