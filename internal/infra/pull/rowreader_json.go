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

package pull

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/cgi-fr/lino/pkg/pull"
)

// JSONRowJSONRowReader read row from JSONLine file
type JSONRowReader struct {
	file     io.Reader
	fscanner *bufio.Scanner
	err      error
	value    *pull.Row
}

// NNewJSONRowReader create a new JSONRowReader
func NewJSONRowReader(file io.Reader) *JSONRowReader {
	return &JSONRowReader{file, bufio.NewScanner(file), nil, nil}
}

// Next return true if Next Value is present
func (jrr *JSONRowReader) Next() bool {
	if jrr.fscanner.Scan() {
		line := jrr.fscanner.Bytes()
		var internalValue pull.Row
		err := json.Unmarshal(line, &internalValue)
		if err != nil {
			jrr.err = err
			return false
		}
		jrr.value = &internalValue
		return true
	}
	if jrr.fscanner.Err() != nil {
		jrr.err = jrr.fscanner.Err()
	}
	return false
}

// Value return the current Row
func (jrr *JSONRowReader) Value() pull.Row {
	if jrr.value != nil {
		return *jrr.value
	}
	panic("Value is not valid after iterator finished")
}

func (jrr *JSONRowReader) Error() error {
	return jrr.err
}
