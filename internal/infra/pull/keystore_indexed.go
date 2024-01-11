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
	"encoding/json"

	"github.com/cgi-fr/lino/pkg/pull"
)

type KeyStoreIndexed struct {
	keys  []string
	store map[string]*KeyStoreIndexed
}

func NewKeyStoreIndexed(keys []string) *KeyStoreIndexed {
	return &KeyStoreIndexed{
		keys:  keys,
		store: map[string]*KeyStoreIndexed{},
	}
}

func (is *KeyStoreIndexed) encode(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}

func (is *KeyStoreIndexed) Add(data pull.Row) {
	if len(is.keys) == 0 {
		return
	}

	value := is.encode(data[is.keys[0]])

	sub, ok := is.store[value]
	if !ok {
		sub = NewKeyStoreIndexed(is.keys[1:])
	}

	sub.Add(data)

	is.store[value] = sub
}

func (is *KeyStoreIndexed) Has(data pull.Row) bool {
	if len(is.keys) == 0 {
		return true
	}

	value := is.encode(data[is.keys[0]])

	sub, ok := is.store[value]
	if !ok {
		return false
	}

	return sub.Has(data)
}
