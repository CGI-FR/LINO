package push

import (
	"encoding/json"
	"io"

	"makeit.imfr.cgi.com/lino/pkg/push"
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
func (rw *JSONRowWriter) Write(row push.Row) *push.Error {
	err := rw.encoder.Encode(row)
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}
