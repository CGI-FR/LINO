// Copyright (C) 2021 CGI France
//
// This file is part of LINO.
//
// LINO is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// LINO is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with LINO.  If not, see <http://www.gnu.org/licenses/>.

package push_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cgi-fr/lino/pkg/push"
	"github.com/stretchr/testify/assert"
)

func makeTable(name string) push.Table {
	return push.NewTable(name, []string{}, nil)
}

func makeRel(from, to push.Table) push.Relation {
	return push.NewRelation(from.Name()+"->"+to.Name(), from, to)
}

func TestSimplePush(t *testing.T) {
	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")

	plan := push.NewPlan(
		A,
		[]push.Relation{},
	)
	ri := rowIterator{limit: 10, row: push.Row{"name": "John"}}
	tables := map[string]*rowWriter{
		A.Name(): {},
		B.Name(): {},
		C.Name(): {},
	}
	dest := memoryDataDestination{tables, false, false, false, 0}

	err := push.Push(&ri, &dest, plan, push.Insert, 2, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, true, dest.closed)
	assert.Equal(t, 10, len(dest.tables[A.Name()].rows))
	assert.Equal(t, "John", dest.tables[A.Name()].rows[0]["name"])

	assert.Equal(t, 0, len(dest.tables[B.Name()].rows))
}

func TestRelationPush(t *testing.T) {
	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")

	AB := makeRel(A, B)
	BC := makeRel(B, C)

	plan := push.NewPlan(
		A,
		[]push.Relation{AB, BC},
	)
	ri := rowIterator{limit: 10, row: push.Row{
		"name": "John",
		"A->B": map[string]interface{}{
			"age": 42,
			"B->C": map[string]interface{}{
				"sex": "M",
			},
		},
	}}

	tables := map[string]*rowWriter{
		A.Name(): {},
		B.Name(): {},
		C.Name(): {},
	}
	dest := memoryDataDestination{tables, false, false, false, 0}

	err := push.Push(&ri, &dest, plan, push.Insert, 2, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	// no error
	assert.Nil(t, err)
	// destination was opened
	assert.Equal(t, true, dest.opened)
	// destination is closed
	assert.Equal(t, true, dest.closed)
	// all rows are inserted table A
	assert.Equal(t, 10, len(dest.tables[A.Name()].rows))
	assert.Equal(t, 1, len(dest.tables[A.Name()].rows[0]))
	assert.Equal(t, "John", dest.tables[A.Name()].rows[0]["name"])
	// all rows are inserted table B
	assert.Equal(t, 10, len(dest.tables[B.Name()].rows))
	assert.Equal(t, 1, len(dest.tables[B.Name()].rows[0]))
	assert.Equal(t, 42, dest.tables[B.Name()].rows[0]["age"])
	// all rows are inserted table C
	assert.Equal(t, 10, len(dest.tables[C.Name()].rows))
	assert.Equal(t, 1, len(dest.tables[C.Name()].rows[0]))
	assert.Equal(t, "M", dest.tables[C.Name()].rows[0]["sex"])
}

func TestRelationPushWithEmptyRelation(t *testing.T) {
	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")

	AB := makeRel(A, B)
	BC := makeRel(B, C)

	plan := push.NewPlan(
		A,
		[]push.Relation{AB, BC},
	)
	ri := rowIterator{limit: 10, row: push.Row{
		"name": "John",
		"A->B": map[string]interface{}{
			"age":  42,
			"B->C": nil,
		},
	}}

	tables := map[string]*rowWriter{
		A.Name(): {},
		B.Name(): {},
		C.Name(): {},
	}
	dest := memoryDataDestination{tables, false, false, false, 0}

	err := push.Push(&ri, &dest, plan, push.Insert, 2, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	// no error
	assert.Nil(t, err)
	// destination was opened
	assert.Equal(t, true, dest.opened)
	// destination is closed
	assert.Equal(t, true, dest.closed)
	// all rows are inserted table A
	assert.Equal(t, 10, len(dest.tables[A.Name()].rows))
	assert.Equal(t, 1, len(dest.tables[A.Name()].rows[0]))
	assert.Equal(t, "John", dest.tables[A.Name()].rows[0]["name"])
	// all rows are inserted table B
	assert.Equal(t, 10, len(dest.tables[B.Name()].rows))
	assert.Equal(t, 1, len(dest.tables[B.Name()].rows[0]))
	assert.Equal(t, 42, dest.tables[B.Name()].rows[0]["age"])
	// No rows are inserted table C
	assert.Equal(t, 0, len(dest.tables[C.Name()].rows))
}

func TestInversseRelationPush(t *testing.T) {
	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")

	AB := makeRel(A, B)
	BC := makeRel(C, B)

	plan := push.NewPlan(
		A,
		[]push.Relation{AB, BC},
	)
	ri := rowIterator{limit: 1, row: push.Row{
		"name": "John",
		"A->B": map[string]interface{}{
			"age": 42,
			"C->B": []interface{}{
				map[string]interface{}{"history": "1"},
				map[string]interface{}{"history": "2"},
				map[string]interface{}{"history": "3"},
			},
		},
	}}

	tables := map[string]*rowWriter{
		A.Name(): {},
		B.Name(): {},
		C.Name(): {},
	}
	dest := memoryDataDestination{tables, false, false, false, 0}

	err := push.Push(&ri, &dest, plan, push.Insert, 5, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	// no error
	assert.Nil(t, err)
	// destination was opened
	assert.Equal(t, true, dest.opened)
	// destination is closed
	assert.Equal(t, true, dest.closed)
	// all rows are inserted table A
	assert.Equal(t, 1, len(dest.tables[A.Name()].rows))
	assert.Equal(t, 1, len(dest.tables[A.Name()].rows[0]))
	assert.Equal(t, "John", dest.tables[A.Name()].rows[0]["name"])
	// all rows are inserted table B
	assert.Equal(t, 1, len(dest.tables[B.Name()].rows))
	assert.Equal(t, 1, len(dest.tables[B.Name()].rows[0]))
	assert.Equal(t, 42, dest.tables[B.Name()].rows[0]["age"])
	// all rows are inserted table C
	assert.Equal(t, 3, len(dest.tables[C.Name()].rows))
	assert.Equal(t, 1, len(dest.tables[C.Name()].rows[0]))
	assert.Equal(t, "1", dest.tables[C.Name()].rows[0]["history"])
}

func TestPushWithCommitTimeout(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})

	// Iterator that returns 2 rows with a delay between them
	ri := &delayedRowIterator{
		rows: []push.Row{
			{"name": "Row1"},
			{"name": "Row2"},
		},
		delay: 200 * time.Millisecond,
	}

	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	// Commit size 10, but timeout 100ms. Should trigger commit after first row due to delay.
	err := push.Push(ri, dest, plan, push.Insert, 10, 100*time.Millisecond, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	// Should have 2 commits: 1 for timeout after first row, 1 final commit for second row
	assert.Equal(t, 2, dest.commits, "Expected 2 commits: 1 timeout + 1 final")
	assert.Equal(t, 2, len(dest.tables[A.Name()].rows))
}

func TestPushWithSavepoint(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "savepoint")
	assert.Nil(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	A := push.NewTable("A", []string{"id"}, nil)
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 5, row: push.Row{"id": 1, "name": "John"}}

	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	// Commit size 2, total 5 rows -> 2 intermediate commits
	err = push.Push(&ri, dest, plan, push.Insert, 2, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", tmpfile.Name(), false)

	assert.Nil(t, err)

	content, err := os.ReadFile(tmpfile.Name())
	assert.Nil(t, err)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")

	assert.GreaterOrEqual(t, len(lines), 4)
}

func TestPushWithObservers(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 5, row: push.Row{"name": "John"}}
	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	obs := &mockObserver{}
	err := push.Push(&ri, dest, plan, push.Insert, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false, obs)

	assert.Nil(t, err)
	assert.Equal(t, 5, obs.pushedCount)
	assert.True(t, obs.closed)
}

type delayedRowIterator struct {
	rows  []push.Row
	index int
	delay time.Duration
}

func (i *delayedRowIterator) Next() bool {
	if i.index >= len(i.rows) {
		return false
	}
	if i.index > 0 {
		time.Sleep(i.delay)
	}
	i.index++
	return true
}

func (i *delayedRowIterator) Value() *push.Row {
	return &i.rows[i.index-1]
}

func (i *delayedRowIterator) Error() *push.Error {
	return nil
}

func (i *delayedRowIterator) Close() *push.Error {
	return nil
}

type mockObserver struct {
	pushedCount int
	closed      bool
}

func (m *mockObserver) Pushed() {
	m.pushedCount++
}

func (m *mockObserver) Close() {
	m.closed = true
}

// Test: Empty iterator (no rows)
func TestPushWithEmptyIterator(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 0, row: push.Row{"name": "John"}}
	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(&ri, dest, plan, push.Insert, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(dest.tables[A.Name()].rows))
	assert.Equal(t, 0, dest.commits, "No commits should occur for empty iterator")
}

// Test: Exactly commitSize rows (boundary condition)
func TestPushWithExactCommitSize(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 5, row: push.Row{"name": "John"}}
	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(&ri, dest, plan, push.Insert, 5, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 5, len(dest.tables[A.Name()].rows))
	// Should have 1 intermediate commit (at row 5) + no final commit since count % commitSize == 0
	assert.Equal(t, 1, dest.commits, "Expected 1 commit for exact commitSize")
}

// Test: commitSize = 1 (commit every row)
func TestPushWithCommitSizeOne(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 3, row: push.Row{"name": "John"}}
	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(&ri, dest, plan, push.Insert, 1, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 3, len(dest.tables[A.Name()].rows))
	assert.Equal(t, 3, dest.commits, "Expected 3 commits (one per row)")
}

// Test: Iterator error handling
func TestPushWithIteratorError(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := &errorRowIterator{
		rows:       []push.Row{{"name": "Row1"}},
		errorAfter: 1,
	}
	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(ri, dest, plan, push.Insert, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.NotNil(t, err)
	assert.Equal(t, "iterator error", err.Description)
}

// Test: Destination open error
func TestPushWithDestinationOpenError(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 5, row: push.Row{"name": "John"}}
	dest := &errorDataDestination{failOnOpen: true}

	err := push.Push(&ri, dest, plan, push.Insert, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.NotNil(t, err)
	assert.Equal(t, "open error", err.Description)
}

// Test: Destination commit error
func TestPushWithCommitError(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 3, row: push.Row{"name": "John"}}
	dest := &errorDataDestination{
		tables:       map[string]*rowWriter{A.Name(): {}},
		failOnCommit: true,
		commitToFail: 1,
	}

	err := push.Push(&ri, dest, plan, push.Insert, 2, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.NotNil(t, err)
	assert.Equal(t, "commit error", err.Description)
}

// Test: Multiple timeouts
func TestPushWithMultipleTimeouts(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})

	ri := &delayedRowIterator{
		rows: []push.Row{
			{"name": "Row1"},
			{"name": "Row2"},
			{"name": "Row3"},
		},
		delay: 150 * time.Millisecond,
	}

	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(ri, dest, plan, push.Insert, 10, 100*time.Millisecond, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 3, len(dest.tables[A.Name()].rows))
	// Should have 2 timeout commits (after row 1 and row 2) + 1 final commit (for row 3)
	assert.Equal(t, 3, dest.commits, "Expected 3 commits: 2 timeouts + 1 final")
}

// Test: Timeout with exact commitSize
func TestPushTimeoutWithExactCommitSize(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})

	ri := &delayedRowIterator{
		rows: []push.Row{
			{"name": "Row1"},
			{"name": "Row2"},
		},
		delay: 150 * time.Millisecond,
	}

	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	// commitSize = 2, so after 2 rows we hit the size limit
	err := push.Push(ri, dest, plan, push.Insert, 2, 100*time.Millisecond, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(dest.tables[A.Name()].rows))
	// First row -> timeout commit, second row -> intermediate commit (size reached)
	assert.Equal(t, 2, dest.commits, "Expected 2 commits: 1 timeout + 1 intermediate")
}

// Helper types for error testing
type errorRowIterator struct {
	rows        []push.Row
	index       int
	errorAfter  int
	failOnClose bool
}

func (i *errorRowIterator) Next() bool {
	if i.index >= len(i.rows) {
		return false
	}
	i.index++
	return true
}

func (i *errorRowIterator) Value() *push.Row {
	return &i.rows[i.index-1]
}

func (i *errorRowIterator) Error() *push.Error {
	if i.index >= i.errorAfter {
		return &push.Error{Description: "iterator error"}
	}
	return nil
}

func (i *errorRowIterator) Close() *push.Error {
	if i.failOnClose {
		return &push.Error{Description: "close error"}
	}
	return nil
}

type errorDataDestination struct {
	tables       map[string]*rowWriter
	failOnOpen   bool
	failOnCommit bool
	failOnClose  bool
	commitToFail int
	commitCount  int
	opened       bool
	closed       bool
}

func (d *errorDataDestination) SafeUrl() string {
	return "mem://error-test"
}

func (d *errorDataDestination) Open(plan push.Plan, mode push.Mode, disableConstraints bool, whereClause string) *push.Error {
	if d.failOnOpen {
		return &push.Error{Description: "open error"}
	}
	d.opened = true
	return nil
}

func (d *errorDataDestination) Commit() *push.Error {
	d.commitCount++
	if d.failOnCommit && d.commitCount >= d.commitToFail {
		return &push.Error{Description: "commit error"}
	}
	return nil
}

func (d *errorDataDestination) Close() *push.Error {
	d.closed = true
	if d.failOnClose {
		return &push.Error{Description: "close error"}
	}
	return nil
}

func (d *errorDataDestination) RowWriter(table push.Table) (push.RowWriter, *push.Error) {
	if d.tables == nil {
		d.tables = map[string]*rowWriter{}
	}
	if _, ok := d.tables[table.Name()]; !ok {
		d.tables[table.Name()] = &rowWriter{}
	}
	return d.tables[table.Name()], nil
}

func (d *errorDataDestination) OpenSQLLogger(string) error {
	return nil
}

// Test: Update mode with translator (covers computeTranslatedKeys)
func TestPushUpdateModeWithTranslator(t *testing.T) {
	A := push.NewTable("A", []string{"id"}, nil)
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 2, row: push.Row{"id": 1, "name": "Updated"}}
	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	// Mock translator
	translator := &mockTranslator{
		translations: map[string]interface{}{
			"A.id.1": 100, // old value for new value 1
		},
	}

	err := push.Push(&ri, dest, plan, push.Update, 10, 0, true, push.NoErrorCaptureRowWriter{}, translator, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(dest.tables[A.Name()].rows))
}

// Test: Delete mode
func TestPushDeleteMode(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 3, row: push.Row{"name": "ToDelete"}}
	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(&ri, dest, plan, push.Delete, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 3, len(dest.tables[A.Name()].rows))
}

// Test: Destination close error
func TestPushWithDestinationCloseError(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 2, row: push.Row{"name": "John"}}
	dest := &errorDataDestination{
		tables:      map[string]*rowWriter{A.Name(): {}},
		failOnClose: true,
	}

	err := push.Push(&ri, dest, plan, push.Insert, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.NotNil(t, err)
	assert.Equal(t, "close error", err.Description)
}

// Test: Iterator close error
func TestPushWithIteratorCloseError(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := &errorRowIterator{
		rows:        []push.Row{{"name": "Row1"}},
		errorAfter:  10, // No error during iteration
		failOnClose: true,
	}
	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(ri, dest, plan, push.Insert, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.NotNil(t, err)
	assert.Equal(t, "close error", err.Description)
}

// Test: Multiple errors (destination close + iterator close)
func TestPushWithMultipleCloseErrors(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := &errorRowIterator{
		rows:        []push.Row{{"name": "Row1"}},
		errorAfter:  10,
		failOnClose: true,
	}
	dest := &errorDataDestination{
		tables:      map[string]*rowWriter{A.Name(): {}},
		failOnClose: true,
	}

	err := push.Push(ri, dest, plan, push.Insert, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.NotNil(t, err)
	assert.Contains(t, err.Description, "close error")
}

// Test: Row write error with catch error writer
func TestPushWithRowWriteError(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 3, row: push.Row{"name": "John"}}

	// Destination with error row writer
	dest := &errorWriterDataDestination{
		tables: map[string]*errorRowWriter{A.Name(): {failAfter: 1}},
	}

	// Catch error writer
	errorWriter := &captureRowWriter{}

	err := push.Push(&ri, dest, plan, push.Insert, 10, 0, true, errorWriter, nil, "", "", "", false)

	assert.Nil(t, err) // Should not fail, errors are caught
	assert.Equal(t, 2, len(errorWriter.rows), "Should have caught 2 errors")
}

// Test: Savepoint write error
func TestPushWithSavepointWriteError(t *testing.T) {
	A := push.NewTable("A", []string{"id"}, nil)
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 3, row: push.Row{"id": 1, "name": "John"}}
	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	// Use invalid path to trigger savepoint error
	err := push.Push(&ri, dest, plan, push.Insert, 2, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "/invalid/path/savepoint.json", "", false)

	assert.Nil(t, err) // Savepoint failure should be non-fatal
}

// Helper types for new tests
type mockTranslator struct {
	translations map[string]interface{}
}

func (m *mockTranslator) FindValue(key push.Key, value push.Value) push.Value {
	keyStr := fmt.Sprintf("%s.%s.%v", key.TableName, key.ColumnName, value)
	if oldValue, ok := m.translations[keyStr]; ok {
		return oldValue
	}
	return value
}

func (m *mockTranslator) Load(keys []push.Key, rows push.RowIterator) *push.Error {
	return nil
}

type errorRowWriter struct {
	rows      []push.Row
	failAfter int
}

func (rw *errorRowWriter) Write(row push.Row, where push.Row) *push.Error {
	if len(rw.rows) >= rw.failAfter {
		return &push.Error{Description: "write error"}
	}
	rw.rows = append(rw.rows, row)
	return nil
}

type captureRowWriter struct {
	rows []push.Row
}

func (c *captureRowWriter) Write(row push.Row, where push.Row) *push.Error {
	c.rows = append(c.rows, row)
	return nil
}

// errorWriterDataDestination is a destination that uses errorRowWriter
type errorWriterDataDestination struct {
	tables map[string]*errorRowWriter
	opened bool
	closed bool
}

func (d *errorWriterDataDestination) SafeUrl() string {
	return "mem://error-writer-test"
}

func (d *errorWriterDataDestination) Open(plan push.Plan, mode push.Mode, disableConstraints bool, whereClause string) *push.Error {
	d.opened = true
	return nil
}

func (d *errorWriterDataDestination) Commit() *push.Error {
	return nil
}

func (d *errorWriterDataDestination) Close() *push.Error {
	d.closed = true
	return nil
}

func (d *errorWriterDataDestination) RowWriter(table push.Table) (push.RowWriter, *push.Error) {
	if d.tables == nil {
		d.tables = map[string]*errorRowWriter{}
	}
	if _, ok := d.tables[table.Name()]; !ok {
		d.tables[table.Name()] = &errorRowWriter{}
	}
	return d.tables[table.Name()], nil
}

func (d *errorWriterDataDestination) OpenSQLLogger(string) error {
	return nil
}

// Test: FilterRelation with invalid array element (not a map)
func TestPushWithInvalidRelationArrayElement(t *testing.T) {
	A := makeTable("A")
	B := makeTable("B")
	AB := makeRel(A, B)

	plan := push.NewPlan(A, []push.Relation{AB})

	// Invalid relation: array element is not a map
	ri := &singleRowIterator{
		row: push.Row{
			"name": "John",
			"A->B": []interface{}{"invalid", "not a map"},
		},
	}

	tables := map[string]*rowWriter{A.Name(): {}, B.Name(): {}}
	dest := &memoryDataDestination{tables: tables}
	errorWriter := &captureRowWriter{}

	err := push.Push(ri, dest, plan, push.Insert, 10, 0, true, errorWriter, nil, "", "", "", false)

	assert.Nil(t, err) // Error should be caught
	assert.Equal(t, 1, len(errorWriter.rows))
}

// Test: FilterRelation with invalid relation type (not map or array)
func TestPushWithInvalidRelationType(t *testing.T) {
	A := makeTable("A")
	B := makeTable("B")
	AB := makeRel(A, B)

	plan := push.NewPlan(A, []push.Relation{AB})

	// Invalid relation: not a map or array
	ri := &singleRowIterator{
		row: push.Row{
			"name": "John",
			"A->B": "invalid string value",
		},
	}

	tables := map[string]*rowWriter{A.Name(): {}, B.Name(): {}}
	dest := &memoryDataDestination{tables: tables}
	errorWriter := &captureRowWriter{}

	err := push.Push(ri, dest, plan, push.Insert, 10, 0, true, errorWriter, nil, "", "", "", false)

	assert.Nil(t, err) // Error should be caught
	assert.Equal(t, 1, len(errorWriter.rows))
}

// Test: Push with whereField
func TestPushWithWhereField(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})

	ri := &singleRowIterator{
		row: push.Row{
			"name": "John",
			"__usingpk__": map[string]interface{}{
				"old_id": 100,
			},
		},
	}

	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(ri, dest, plan, push.Update, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "__usingpk__", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(dest.tables[A.Name()].rows))
}

// Test: Push with table columns (import="no")
func TestPushWithNoImportColumns(t *testing.T) {
	// Create table with columns
	columns := []push.Column{
		push.NewColumn("id", "yes", "yes", 0, false, false, ""),
		push.NewColumn("name", "yes", "yes", 0, false, false, ""),
		push.NewColumn("internal", "yes", "no", 0, false, false, ""), // Should not be imported
	}
	A := push.NewTable("A", []string{"id"}, push.NewColumnList(columns))
	plan := push.NewPlan(A, []push.Relation{})

	ri := &singleRowIterator{
		row: push.Row{
			"id":       1,
			"name":     "John",
			"internal": "should_be_removed",
		},
	}

	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(ri, dest, plan, push.Insert, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(dest.tables[A.Name()].rows))
	// The "internal" field should have been removed
	_, hasInternal := dest.tables[A.Name()].rows[0]["internal"]
	assert.False(t, hasInternal, "internal field should be removed")
}

// Test: Truncate mode
func TestPushTruncateMode(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 2, row: push.Row{"name": "John"}}
	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(&ri, dest, plan, push.Truncate, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(dest.tables[A.Name()].rows))
}

// Test: Update mode with relations
func TestPushUpdateModeWithRelations(t *testing.T) {
	A := makeTable("A")
	B := makeTable("B")
	AB := makeRel(A, B)

	plan := push.NewPlan(A, []push.Relation{AB})

	ri := &singleRowIterator{
		row: push.Row{
			"name": "John",
			"A->B": map[string]interface{}{
				"age": 42,
			},
		},
	}

	tables := map[string]*rowWriter{A.Name(): {}, B.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(ri, dest, plan, push.Update, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(dest.tables[A.Name()].rows))
	assert.Equal(t, 1, len(dest.tables[B.Name()].rows))
}

// Test: Delete mode with relations
func TestPushDeleteModeWithRelations(t *testing.T) {
	A := makeTable("A")
	B := makeTable("B")
	AB := makeRel(A, B)

	plan := push.NewPlan(A, []push.Relation{AB})

	ri := &singleRowIterator{
		row: push.Row{
			"name": "John",
			"A->B": map[string]interface{}{
				"age": 42,
			},
		},
	}

	tables := map[string]*rowWriter{A.Name(): {}, B.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(ri, dest, plan, push.Delete, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(dest.tables[A.Name()].rows))
	assert.Equal(t, 1, len(dest.tables[B.Name()].rows))
}

// Test: Update mode with inverse relations
func TestPushUpdateModeWithInverseRelations(t *testing.T) {
	A := makeTable("A")
	C := makeTable("C")
	B := makeTable("B")
	AB := makeRel(A, B)
	CB := makeRel(C, B)

	plan := push.NewPlan(A, []push.Relation{AB, CB})

	ri := &singleRowIterator{
		row: push.Row{
			"name": "John",
			"A->B": map[string]interface{}{
				"age": 42,
				"C->B": []interface{}{
					map[string]interface{}{"history": "1"},
				},
			},
		},
	}

	tables := map[string]*rowWriter{A.Name(): {}, B.Name(): {}, C.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	err := push.Push(ri, dest, plan, push.Update, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(dest.tables[A.Name()].rows))
	assert.Equal(t, 1, len(dest.tables[B.Name()].rows))
	assert.Equal(t, 1, len(dest.tables[C.Name()].rows))
}

// Helper: single row iterator
type singleRowIterator struct {
	row  push.Row
	done bool
}

func (i *singleRowIterator) Next() bool {
	if i.done {
		return false
	}
	i.done = true
	return true
}

func (i *singleRowIterator) Value() *push.Row {
	return &i.row
}

func (i *singleRowIterator) Error() *push.Error {
	return nil
}

func (i *singleRowIterator) Close() *push.Error {
	return nil
}

// Test: Stats computation
func TestStatsComputation(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 5, row: push.Row{"name": "John"}}
	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	push.Reset()

	err := push.Push(&ri, dest, plan, push.Insert, 2, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)

	stats := push.Compute()
	assert.Equal(t, 5, stats.GetInputLinesCount())
	assert.Equal(t, 5, stats.GetCreatedLinesCount()[A.Name()])
	assert.Equal(t, 3, stats.GetCommitsCount()) // 2 intermediate + 1 final

	// Test ToJSON
	jsonBytes := stats.ToJSON()
	assert.NotNil(t, jsonBytes)
	assert.Greater(t, len(jsonBytes), 0)
}

// Test: Delete stats
func TestDeleteStats(t *testing.T) {
	A := makeTable("A")
	plan := push.NewPlan(A, []push.Relation{})
	ri := rowIterator{limit: 3, row: push.Row{"name": "John"}}
	tables := map[string]*rowWriter{A.Name(): {}}
	dest := &memoryDataDestination{tables: tables}

	push.Reset()

	err := push.Push(&ri, dest, plan, push.Delete, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", "", false)

	assert.Nil(t, err)

	stats := push.Compute()
	assert.Equal(t, 3, stats.GetDeletedLinesCount()[A.Name()])
}

// Test: Mode parsing
func TestModeParsing(t *testing.T) {
	mode, err := push.ParseMode("insert")
	assert.Nil(t, err)
	assert.Equal(t, push.Insert, mode)
	assert.Equal(t, "insert", mode.String())

	mode, err = push.ParseMode("update")
	assert.Nil(t, err)
	assert.Equal(t, push.Update, mode)

	mode, err = push.ParseMode("delete")
	assert.Nil(t, err)
	assert.Equal(t, push.Delete, mode)

	mode, err = push.ParseMode("truncate")
	assert.Nil(t, err)
	assert.Equal(t, push.Truncate, mode)

	_, err = push.ParseMode("invalid")
	assert.NotNil(t, err)
}

// Test: Plan Tables
func TestPlanTables(t *testing.T) {
	A := makeTable("A")
	B := makeTable("B")
	AB := makeRel(A, B)

	plan := push.NewPlan(A, []push.Relation{AB})

	tables := plan.Tables()
	assert.Equal(t, 2, len(tables))
}

// Test: Column methods
func TestColumnMethods(t *testing.T) {
	col := push.NewColumn("test", "yes", "yes", 100, false, true, "preserve_value")

	assert.Equal(t, "test", col.Name())
	assert.Equal(t, "yes", col.Export())
	assert.Equal(t, int64(100), col.Length())
	assert.False(t, col.LengthInBytes())
	assert.True(t, col.Truncate())
	assert.Equal(t, "preserve_value", col.Preserve())
}

// Test: Table GetColumn
func TestTableGetColumn(t *testing.T) {
	columns := []push.Column{
		push.NewColumn("id", "yes", "yes", 0, false, false, ""),
		push.NewColumn("name", "yes", "yes", 0, false, false, ""),
	}
	A := push.NewTable("A", []string{"id"}, push.NewColumnList(columns))

	col := A.GetColumn("name")
	assert.NotNil(t, col)
	assert.Equal(t, "name", col.Name())

	col = A.GetColumn("nonexistent")
	assert.Nil(t, col)
}
