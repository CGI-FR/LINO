package relation

import (
	"fmt"

	"github.com/cgi-fr/lino/pkg/relation"

	// import Oracle connector
	_ "github.com/sijms/go-ora/v2"
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
	if schema == "" {
		schema = "user"
	} else {
		schema = fmt.Sprintf("'%s'", schema)
	}

	SQL := `
SELECT
    fk.constraint_name    AS relation_name,
    fk.table_name         AS child_table,
    fk_cols.column_name   AS child_column,
    pk.table_name         AS parent_table,
    pk_cols.column_name   AS parent_column
FROM
    all_constraints fk
-- 1. Récupérer le parent (en précisant le R_OWNER car le parent peut être dans un autre schéma)
JOIN
    all_constraints pk 
    ON fk.r_constraint_name = pk.constraint_name 
    AND fk.r_owner = pk.owner
-- 2. Colonnes de l'enfant (Jointure sur NOM + OWNER)
JOIN
    all_cons_columns fk_cols 
    ON fk.constraint_name = fk_cols.constraint_name 
    AND fk.owner = fk_cols.owner
-- 3. Colonnes du parent (Jointure sur NOM + OWNER)
JOIN
    all_cons_columns pk_cols 
    ON pk.constraint_name = pk_cols.constraint_name 
    AND pk.owner = pk_cols.owner
WHERE
    fk.constraint_type = 'R' -- foreignkey
    AND fk_cols.position = pk_cols.position
    -- FILTRE SUR LE PROPRIÉTAIRE ICI :
    AND fk.owner = ` + schema + `
ORDER BY
    fk.table_name,
    fk.constraint_name,
    fk_cols.position
`

	return SQL
}
