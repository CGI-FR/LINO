package push_test

import (
	"fmt"

	"makeit.imfr.cgi.com/lino/pkg/push"
)

type rowIterator struct {
	limit uint
	row   push.Row
}

func (ri *rowIterator) NextRow() (*push.Row, *push.StopIteratorError) {
	if ri.limit == 0 {
		return nil, &push.StopIteratorError{}
	}
	ri.limit--
	return &ri.row, nil
}

func (ri *rowIterator) Close() *push.Error {
	return nil
}

type memoryDataDestination struct {
	tables map[string]*rowWriter
	closed bool
	opened bool
}

func (mdd *memoryDataDestination) RowWriter(table push.Table) (push.RowWriter, *push.Error) {
	return mdd.tables[table.Name()], nil
}

func (mdd *memoryDataDestination) Open(pla push.Plan, mode push.Mode) *push.Error {
	mdd.opened = true
	return nil
}

func (mdd *memoryDataDestination) Close() *push.Error {
	mdd.closed = true
	return nil
}

type rowWriter struct {
	rows []push.Row
}

func (rw *rowWriter) Write(row push.Row) *push.Error {
	logger.Trace(fmt.Sprintf("append row %s to %s", row, rw.rows))
	rw.rows = append(rw.rows, row)
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
