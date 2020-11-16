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
	err      *pull.Error
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
			jrr.err = &pull.Error{Description: err.Error()}
			return false
		}
		jrr.value = &internalValue
		return true
	}
	if jrr.fscanner.Err() != nil {
		jrr.err = &pull.Error{Description: jrr.fscanner.Err().Error()}
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

func (jrr *JSONRowReader) Error() *pull.Error {
	return jrr.err
}
