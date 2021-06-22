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

// +build db2

package relation

import (
	"fmt"

	// import db2 connector
	_ "github.com/ibmdb/go_ibm_db"

	"github.com/cgi-fr/lino/pkg/relation"
)

// NewDb2ExtractorFactory creates a new db2 extractor factory.
func NewDb2ExtractorFactory() *Db2ExtractorFactory {
	return &Db2ExtractorFactory{}
}

// Db2ExtractorFactory exposes methods to create new Db2 extractors.
type Db2ExtractorFactory struct{}

// New return a Db2 extractor
func (e *Db2ExtractorFactory) New(url string, schema string) relation.Extractor {
	return NewSQLExtractor(url, schema, Db2Dialect{})
}

type Db2Dialect struct{}

func (d Db2Dialect) SQL(schema string) string {
	SQL := `
	select
		ref.constname as constraint_name,
		ref.reftabname as primary_table,
		ref.refkeyname as primary_key,
		ref.tabname as foreign_table,
		LISTAGG(k.colname, ',') AS foreign_key_names
	from syscat.references ref
	inner join syscat.keycoluse k
		on ref.constname = k.constname
	where ref.ownertype = 'U'
	group by ref.constname, ref.reftabname, ref.refkeyname, ref.tabname
  `

	if schema != "" {
		SQL += fmt.Sprintf(" and ref.tabschema = '%s' and ref.reftabschema = '%s'", schema, schema)
	}
	return SQL
}
