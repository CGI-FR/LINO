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
	"encoding/json"
	"fmt"
	"io"

	"github.com/cgi-fr/lino/pkg/pull"
)

// JSONRowExporter export rows to JSON format.
type JSONRowExporter struct {
	file io.Writer
}

// NewJSONRowExporter creates a new JSONRowExporter.
func NewJSONRowExporter(file io.Writer) *JSONRowExporter {
	return &JSONRowExporter{file}
}

// Export rows in JSON format.
func (re *JSONRowExporter) Export(r pull.ExportedRow) error {
	jsonString, err := json.Marshal(r)
	if err != nil {
		return err
	}
	fmt.Fprintln(re.file, string(jsonString))
	return nil
}
