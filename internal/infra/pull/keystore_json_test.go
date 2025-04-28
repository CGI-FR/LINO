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
	"bytes"
	"strings"
	"testing"

	"github.com/cgi-fr/lino/pkg/pull"
	"github.com/stretchr/testify/assert"
)

func TestJSONKeyStore(t *testing.T) {
	var buffer bytes.Buffer
	buffer.WriteString(`{"ID_INDIVIDU":1000006456,"ID_LANGUE":472355}`)

	ks, err := NewJSONKeyStore(&buffer, []string{"ID_INDIVIDU", "ID_LANGUE"})

	assert.Nil(t, err)
	assert.True(t, ks.Has(pull.Row{"ID_INDIVIDU": 1000006456, "ID_LANGUE": 472355}))

	assert.True(t, ks.Has(pull.Row{"ID_INDIVIDU": 1000006456, "ID_LANGUE": 472355, "OTHER": 42}))

	assert.False(t, ks.Has(pull.Row{"ID_INDIVIDU": 1000008957, "ID_LANGUE": 472355}))
	assert.False(t, ks.Has(pull.Row{"ID_INDIVIDU": 1000006456, "ID_LANGUE": 472354, "OTHER": 42}))
	assert.False(t, ks.Has(pull.Row{"ID_INDIVIDU": 1000006456, "OTHER": 42}))
}

func TestJSONKeyStoreFromJSON(t *testing.T) {
	ks, err := NewJSONKeyStore(strings.NewReader(
		`{"ID_INDIVIDU":1000006456,"ID_LANGUE":472355}`),
		[]string{"ID_INDIVIDU", "ID_LANGUE"})

	assert.Nil(t, err)

	jrr := NewJSONRowReader(strings.NewReader(`{"ID_INDIVIDU":1000006456,"ID_LANGUE":472355,"CODE_LANGUE":"AN","NIVEAU":"2","COMPLEMENT":null,"JSON_LIAISON":null,"ORIGINE_DONNEE":"MIGAUDEPP","VISIBILITE_DONNEE":"C"}`))

	assert.True(t, jrr.Next())

	row := jrr.Value()
	assert.True(t, ks.Has(row))
}
