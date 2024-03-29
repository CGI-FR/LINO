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

package main

import (
	infra "github.com/cgi-fr/lino/internal/infra/analyse"
)

func analyseDataSourceFactory() map[string]infra.SQLExtractorFactory {
	return map[string]infra.SQLExtractorFactory{
		"postgres":   infra.NewPostgresExtractorFactory(),
		"godror":     infra.NewOracleExtractorFactory(),
		"godror-raw": infra.NewOracleExtractorFactory(),
		"mysql":      infra.NewMariaDBExtractorFactory(),
		"db2":        infra.NewDB2ExtractorFactory(),
		"sqlserver":  infra.NewSQLServerExtractorFactory(),
	}
}
