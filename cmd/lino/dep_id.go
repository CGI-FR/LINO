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
	"os"

	infra "makeit.imfr.cgi.com/lino/internal/infra/id"
	domain "makeit.imfr.cgi.com/lino/pkg/id"
)

func idStorage() domain.Storage {
	return infra.NewMultiStorage(infra.NewYAMLStorage(), infra.NewDOTStorage())
}

func idStorageFactory() func(string) domain.Storage {
	return func(table string) domain.Storage {
		if table == "" {
			return idStorage()
		}
		return infra.NewTableStorage(domain.NewTable(table))
	}
}

func idExporter() domain.Exporter {
	return infra.NewGraphVizExporter()
}

func idJSONStorage(file os.File) domain.Storage {
	return infra.NewJSONStorage(file)
}
