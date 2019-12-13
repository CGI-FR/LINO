package load

import (
	"bufio"
	"encoding/json"
	"os"

	"makeit.imfr.cgi.com/lino/pkg/load"
)

// JSONRowIterator export rows to JSON format.
type JSONRowIterator struct {
	file     *os.File
	fscanner *bufio.Scanner
}

// NewJSONRowIterator creates a new JSONRowIterator.
func NewJSONRowIterator(file *os.File) load.RowIterator {
	return &JSONRowIterator{file, bufio.NewScanner(file)}
}

// Close file format.
func (re *JSONRowIterator) Close() *load.Error {
	err := re.file.Close()
	if err != nil {
		return &load.Error{Description: err.Error()}
	}
	return nil
}

// NextRow convert next line to Row
func (re *JSONRowIterator) NextRow() (*load.Row, *load.StopIteratorError) {
	if !re.fscanner.Scan() {
		return nil, &load.StopIteratorError{}
	}
	line := re.fscanner.Bytes()

	var row load.Row
	err2 := json.Unmarshal(line, &row)
	if err2 != nil {
		return nil, &load.StopIteratorError{}
	}
	return &row, nil
}
