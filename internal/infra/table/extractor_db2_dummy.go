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

//go:build !db2
// +build !db2

package table

import (
	"fmt"

	"github.com/cgi-fr/lino/internal/infra/commonsql"
	"github.com/cgi-fr/lino/pkg/table"
)

// NewDb2ExtractorFactory creates a new postgres extractor factory.
func NewDb2ExtractorFactory() *Db2ExtractorFactory {
	return &Db2ExtractorFactory{}
}

// Db2ExtractorFactory exposes methods to create new Db2 extractors.
type Db2ExtractorFactory struct{}

// New return a Db2 extractor
func (e *Db2ExtractorFactory) New(url string, schema string) table.Extractor {
	return NewSQLExtractor(url, schema, Db2Dialect{commonsql.Db2Dialect{}})
}

type Db2Dialect struct {
	commonsql.Dialect
}

func (d Db2Dialect) SQL(schema string) string {
	panic(fmt.Errorf("Not implemented"))
}

func (d Db2Dialect) GetExportType(dbtype string) (string, bool) {
	return "", false
}
