package push

import (
	"bufio"
	"encoding/json"
	"os"

	"makeit.imfr.cgi.com/lino/pkg/push"
)

// JSONRowIterator export rows to JSON format.
type JSONRowIterator struct {
	file     *os.File
	fscanner *bufio.Scanner
}

// NewJSONRowIterator creates a new JSONRowIterator.
func NewJSONRowIterator(file *os.File) push.RowIterator {
	return &JSONRowIterator{file, bufio.NewScanner(file)}
}

// Close file format.
func (re *JSONRowIterator) Close() *push.Error {
	err := re.file.Close()
	if err != nil {
		return &push.Error{Description: err.Error()}
	}
	return nil
}

// NextRow convert next line to Row
func (re *JSONRowIterator) NextRow() (*push.Row, *push.StopIteratorError) {
	if !re.fscanner.Scan() {
		return nil, &push.StopIteratorError{}
	}
	line := re.fscanner.Bytes()

	var row push.Row
	err2 := json.Unmarshal(line, &row)
	if err2 != nil {
		return nil, &push.StopIteratorError{}
	}
	return &row, nil
}
