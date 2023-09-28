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

package analyse

// DataSource is the provider of data to analyse
type DataSource interface {
	BaseName() string
	Next() bool
	Value() ([]interface{}, string, string, error)
}

// Analyser is the provider of statistics analyse
type Analyser interface {
	Analyse(ds DataSource) error
}

// Reporter is the provider of reporting result of the analyse
type Reporter interface {
	Export() error
}
