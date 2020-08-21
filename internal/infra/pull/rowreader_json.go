package pull

import (
	"bufio"
	"encoding/json"
	"io"

	"makeit.imfr.cgi.com/lino/pkg/pull"
)

// JSONRowJSONRowReader read row from JSONLine file
type JSONRowReader struct {
	file     io.Reader
	fscanner *bufio.Scanner
}

// NNewJSONRowReader create a new JSONRowReader
func NewJSONRowReader(file io.ReadCloser) JSONRowReader {
	return JSONRowReader{file, bufio.NewScanner(file)}
}

// Next return true if Next Value is present
func (jrr JSONRowReader) Next() bool {
	return jrr.fscanner.Scan()
}

// Value return the current Row
func (jrr JSONRowReader) Value() (pull.Row, *pull.Error) {
	line := jrr.fscanner.Bytes()
	var internalValue pull.Row
	err := json.Unmarshal(line, &internalValue)
	if err != nil {
		return internalValue, &pull.Error{Description: err.Error()}
	}
	return internalValue, nil
}
