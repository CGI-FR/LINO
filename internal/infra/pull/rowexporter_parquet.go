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
	"github.com/parquet-go/parquet-go"
)

// ParquetRowExporter export rows to Parquet format.
type ParquetRowExporter struct {
	file io.Writer
}

// NewParquetRowExporter creates a new ParquetRowExporter.
func NewParquetRowExporter(file io.Writer) *ParquetRowExporter {
	return &ParquetRowExporter{file}
}

// Export rows in JSON format.
func (re *ParquetRowExporter) Export(r pull.ExportedRow) error {
	parquet.NewGenericWriter
	jsonString, err := json.Marshal(r)
	if err != nil {
		return err
	}
	fmt.Fprintln(re.file, string(jsonString))
	return nil
}
