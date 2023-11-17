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
	"fmt"
	"log"

	_ "github.com/microsoft/go-mssqldb"

	"github.com/cgi-fr/lino/pkg/relation"
)

// NewSQLServerExtractorFactory creates a new SQL Server extractor factory.
func NewSQLServerExtractorFactory() *SQLServerExtractorFactory {
	return &SQLServerExtractorFactory{}
}

// SQLServerExtractorFactory exposes methods to create new SQL Server extractors.
type SQLServerExtractorFactory struct{}

// New return a SQL Server extractor
func (e *SQLServerExtractorFactory) New(connectionString string, schema string) relation.Extractor {
	return NewSQLExtractor(connectionString, schema, SQLServerDialect{})
}

type SQLServerDialect struct{}

func (d SQLServerDialect) SQL(schema string) string {
	SQL := `
	-- Requête principale pour récupérer les informations sur les clés étrangères
	SELECT
		'FK__' + TP2.name + '__' + TP.name + '__' + CONVERT(VARCHAR(40), FK.object_id) AS name,
		child.name AS child_name,
		child.keys,
		parent.name AS parent_name,
		parent.keys
	FROM
		sys.foreign_keys FK
	JOIN
		sys.tables TP ON FK.parent_object_id = TP.object_id
	JOIN
		sys.tables TP2 ON FK.referenced_object_id = TP2.object_id
	-- Pour avoir les colonnes parents on fais une jointure avec la sous requetes
	JOIN
		(
			SELECT
				FK.name AS constraint_name,
				TP2.name AS name,
				STRING_AGG(ccu.column_name, ',') AS keys
			FROM
				sys.foreign_keys FK
			JOIN
				sys.tables TP2 ON FK.referenced_object_id = TP2.object_id
			JOIN
				information_schema.constraint_column_usage AS ccu
				ON FK.name = ccu.constraint_name
			WHERE
				FK.type = 'F'
			GROUP BY
				FK.name, TP2.name
		) parent
		ON FK.name = parent.constraint_name AND TP2.name = parent.name

	-- Pour avoir les colonnes child on fais une jointure avec la sous requetes
	JOIN
		(
			SELECT
				FK.name AS constraint_name,
				TP.name AS name,
				STRING_AGG(kcu.column_name, ',') AS keys
			FROM
				sys.foreign_keys FK
			JOIN
				sys.tables TP ON FK.parent_object_id = TP.object_id
			JOIN
				information_schema.key_column_usage AS kcu
				ON FK.name = kcu.constraint_name
			WHERE
				FK.type = 'F'
			GROUP BY
				FK.name, TP.name
		) child
		ON FK.name = child.constraint_name AND TP.name = child.name;

`
	// Ajouter des logs pour déboguer
	log.Println("Generated SQL:")
	log.Println(SQL)

	if schema != "" {
		SQL += fmt.Sprintf("AND tc.table_schema = '%s'", schema)
	}
	return SQL
}
