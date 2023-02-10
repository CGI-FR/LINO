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

package push

import (
	"encoding/json"
	"io"

	"github.com/cgi-fr/lino/pkg/push"
)

// JSONRowWriter export rows to JSON format.
type JSONRowWriter struct {
	encoder *json.Encoder
}

// NewJSONRowWriter creates a new JSONRowWriter.
func NewJSONRowWriter(file io.Writer) push.RowWriter {
	return &JSONRowWriter{json.NewEncoder(file)}
}

// NextRow convert next line to Row
func (rw *JSONRowWriter) Write(row push.Row, where push.Row) *push.Error {
	err := rw.encoder.Encode(row)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}
