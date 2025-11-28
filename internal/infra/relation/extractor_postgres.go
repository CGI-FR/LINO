// Copyright (C) 2021 CGI France
//
// This file is part of LINO.
//
// LINO is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// LINO is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with LINO.  If not, see <http://www.gnu.org/licenses/>.

package relation

import (

	// import postgresql connector

	_ "github.com/lib/pq"

	"github.com/cgi-fr/lino/pkg/relation"
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
    fk.conname AS relation_name,
    child.relname AS child_table,
    att_child.attname AS child_column,
    parent.relname AS parent_table,
    att_parent.attname AS parent_column
FROM
    pg_constraint fk
    -- Jointure sur le schéma parent et enfant
JOIN
    pg_class child ON fk.conrelid = child.oid
JOIN
    pg_class parent ON fk.confrelid = parent.oid
    -- Jointure pour l'alignement des colonnes (index correspond à index)
JOIN
    unnest(fk.conkey) WITH ORDINALITY AS att_child_pos (attid, pos)
    ON att_child_pos.pos <= array_length(fk.conkey, 1)
JOIN
    unnest(fk.confkey) WITH ORDINALITY AS att_parent_pos (attid, pos)
    ON att_parent_pos.pos = att_child_pos.pos
    -- Récupération du nom des colonnes enfants
JOIN
    pg_attribute att_child ON att_child.attrelid = child.oid
    AND att_child.attnum = att_child_pos.attid
    -- Récupération du nom des colonnes parentes
JOIN
    pg_attribute att_parent ON att_parent.attrelid = parent.oid
    AND att_parent.attnum = att_parent_pos.attid
WHERE
    fk.contype = 'f' -- 'f' pour Foreign Key
`
	if schema != "" {
		SQL += `AND child.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = '` + schema + `')`
	}

	SQL += `
ORDER BY
    child.relname,
    fk.conname,
    att_child_pos.pos
`

	return SQL
}
