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

package table

// ExtractorFactory exposes methods to create new extractors.
type ExtractorFactory interface {
	New(url string, schema string) Extractor
}

// Extractor allows to extract primary keys from a relational database.
type Extractor interface {
	Extract() ([]Table, *Error)
	Count(tableName string) (int, *Error)
}

// Storage allows to store and retrieve Tables objects.
type Storage interface {
	List() ([]Table, *Error)
	Store(tables []Table) *Error
}
