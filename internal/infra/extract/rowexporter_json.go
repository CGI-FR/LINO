package extract

import (
	"encoding/json"
	"fmt"
	"os"

	"makeit.imfr.cgi.com/lino/pkg/extract"
)

// JSONRowExporter export rows to JSON format.
type JSONRowExporter struct {
	file *os.File
}

// NewJSONRowExporter creates a new JSONRowExporter.
func NewJSONRowExporter(file *os.File) *JSONRowExporter {
	return &JSONRowExporter{file}
}

// Export rows in JSON format.
func (re *JSONRowExporter) Export(r extract.Row) *extract.Error {
	jsonString, err := json.Marshal(r)
	if err != nil {
		return &extract.Error{Description: err.Error()}
	}
	fmt.Fprintln(re.file, string(jsonString))
	return nil
}
