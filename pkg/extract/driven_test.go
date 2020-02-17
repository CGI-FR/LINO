package extract_test

import (
	"fmt"

	"makeit.imfr.cgi.com/lino/pkg/extract"
)

// MemoryDataIterator
type MemoryDataIterator struct {
	rows    []extract.Row
	current extract.Row
}

func (di *MemoryDataIterator) HasNext() bool {
	return len(di.rows) > 0
}

func (di *MemoryDataIterator) Next() bool {
	if di.HasNext() {
		di.current = di.rows[0]
		di.rows = di.rows[1:]
		return true
	}

	return false
}

func (di *MemoryDataIterator) Value() (extract.Row, *extract.Error) {
	return di.current, nil
}

// MemoryDataSource mocks DataSource.
type MemoryDataSource struct {
	data map[string][]extract.Row
}

func copyRow(r extract.Row) extract.Row {
	copy := extract.Row{}
	for key, value := range r {
		copy[key] = value
	}
	return copy
}

func (ds *MemoryDataSource) Open() *extract.Error {
	return nil
}

func (ds *MemoryDataSource) Close() *extract.Error {
	return nil
}

func (ds *MemoryDataSource) RowReader(source extract.Table, filter extract.Filter) (extract.RowReader, *extract.Error) {
	rows, ok := ds.data[source.Name()]
	result := []extract.Row{}
	if ok {
	loop:
		for _, row := range rows {
			if len(filter.Values()) == 0 {
				result = append(result, copyRow(row))
				if uint(len(result)) == filter.Limit() {
					break
				}
				continue
			}
			for key, value := range filter.Values() {
				if row[key] == value {
					result = append(result, copyRow(row))
					if uint(len(result)) == filter.Limit() {
						break loop
					}
					continue loop
				}
			}
		}
	}
	fmt.Println("SELECT FROM", source.Name(), "WHERE", filter, "\n  returned:", result)
	return &MemoryDataIterator{result, nil}, nil
}

// MemoryRowExporter mock RowExporter.
type MemoryRowExporter struct {
	rows []extract.Row
}

func (re *MemoryRowExporter) Export(r extract.Row) *extract.Error {
	re.rows = append(re.rows, r)
	fmt.Println("Exported:", r)
	return nil
}

// Logger implementation.
type Logger struct{}

// Trace event.
func (l Logger) Trace(msg string) {
	fmt.Printf("[trace] %v\n", msg)
}

// Debug event.
func (l Logger) Debug(msg string) {
	fmt.Printf("[debug] %v\n", msg)
}

// Info event.
func (l Logger) Info(msg string) {
	fmt.Printf("[info]  %v\n", msg)
}

// Warn event.
func (l Logger) Warn(msg string) {
	fmt.Printf("[warn]  %v\n", msg)
}

// Error event.
func (l Logger) Error(msg string) {
	fmt.Printf("[error] %v\n", msg)
}
