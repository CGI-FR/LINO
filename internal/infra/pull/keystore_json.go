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

package pull

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/cgi-fr/lino/pkg/pull"
)

// JSONKeyStore read row from JSONLine file
type JSONKeyStore struct {
	store map[string]interface{}
}

// NewJSONKeyStore create a new JSONKeyStore
func NewJSONKeyStore(file io.Reader) *JSONKeyStore {
	reader := &JSONRowReader{file, bufio.NewScanner(file), nil, nil}
	store := map[string]interface{}{}
	for reader.Next() {
		bytes, err := json.Marshal(reader.Value())
		if err != nil {
			panic(err)
		}

		store[string(bytes)] = true
	}

	return &JSONKeyStore{
		store: store,
	}
}

func (ks *JSONKeyStore) Has(row pull.Row) bool {
	bytes, err := json.Marshal(row)
	if err != nil {
		panic(err)
	}

	_, ok := ks.store[string(bytes)]

	return ok
}
