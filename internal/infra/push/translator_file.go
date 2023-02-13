// Copyright (C) 2023 CGI France
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

package push

import (
	"github.com/cgi-fr/lino/pkg/push"
)

type FileTranslator struct {
	caches map[string]push.Cache
}

func NewFileTranslator() *FileTranslator {
	return &FileTranslator{caches: map[string]push.Cache{}}
}

func (ft *FileTranslator) Load(tableName string, columnName string, rows push.RowIterator) *push.Error {
	cache := push.Cache{}
	ft.caches[tableName+"___"+columnName] = cache

	for rows.Next() {
		row := *rows.Value()
		cache[row["value"]] = row["key"]
	}

	return rows.Error()
}

func (ft *FileTranslator) FindValue(tableName string, columnName string, value push.Value) push.Value {
	if cache, exists := ft.caches[tableName+"___"+columnName]; exists {
		if oldvalue, exists := cache[value]; exists {
			return oldvalue
		}
	}
	return value
}
