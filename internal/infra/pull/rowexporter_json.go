package pull

import (
	"encoding/json"
	"fmt"
	"io"

	"makeit.imfr.cgi.com/lino/pkg/pull"
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
func (re *JSONRowExporter) Export(r pull.Row) *pull.Error {
	jsonString, err := json.Marshal(r)
	if err != nil {
		return &pull.Error{Description: err.Error()}
	}
	fmt.Fprintln(re.file, string(jsonString))
	return nil
}
