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
	"io"

	infra "github.com/cgi-fr/lino/internal/infra/push"
	domain "github.com/cgi-fr/lino/pkg/push"
)

func pushDataDestinationFactory() map[string]domain.DataDestinationFactory {
	return map[string]domain.DataDestinationFactory{
		"postgres":   infra.NewPostgresDataDestinationFactory(),
		"godror":     infra.NewOracleDataDestinationFactory(),
		"godror-raw": infra.NewOracleDataDestinationFactory(),
		"db2":        infra.NewDb2DataDestinationFactory(),
		"http":       infra.NewHTTPDataDestinationFactory(),
	}
}

func pushRowIteratorFactory() func(io.ReadCloser) domain.RowIterator {
	return infra.NewJSONRowIterator
}

func pushRowExporterFactory() func(io.Writer) domain.RowWriter {
	return infra.NewJSONRowWriter
}
