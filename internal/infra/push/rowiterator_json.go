package push

import (
	"bufio"
	"encoding/json"
	"io"

	"makeit.imfr.cgi.com/lino/pkg/push"
)

// JSONRowIterator export rows to JSON format.
type JSONRowIterator struct {
	file     io.ReadCloser
	fscanner *bufio.Scanner
	error    *push.Error
	value    *push.Row
}

// NewJSONRowIterator creates a new JSONRowIterator.
func NewJSONRowIterator(file io.ReadCloser) push.RowIterator {
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	return &JSONRowIterator{file, scanner, nil, nil}
}

// Close file format.
func (re *JSONRowIterator) Close() *push.Error {
	err := re.file.Close()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

// Value return current row
func (re *JSONRowIterator) Value() *push.Row {
	if re.value != nil {
		return re.value
	}
	panic("Value is not valid after iterator finished")
}

// Error return error catch by next
func (re *JSONRowIterator) Error() *push.Error {
	return re.error
}

// Next try to convert next line to Row
func (re *JSONRowIterator) Next() bool {
	if !re.fscanner.Scan() {
		if re.fscanner.Err() != nil {
			re.error = &push.Error{Description: re.fscanner.Err().Error()}
		}
		return false
	}
	line := re.fscanner.Bytes()

	var row push.Row

	err2 := json.Unmarshal(line, &row)

	if err2 != nil {
		re.error = &push.Error{Description: err2.Error()}
		return false
	}

	re.value = &row

	return true
}
