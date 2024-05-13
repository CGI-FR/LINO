package query

import (
	"encoding/json"
	"fmt"
	"io"
)

type JSONWriter struct {
	file io.Writer
}

// NewJSONWriter creates a new JSONWriter.
func NewJSONWriter(file io.Writer) *JSONWriter {
	return &JSONWriter{file}
}

func (w *JSONWriter) Write(row any) error {
	jsonString, err := json.Marshal(row)
	if err != nil {
		return err
	}
	fmt.Fprintln(w.file, string(jsonString))
	return nil
}
