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

	err := push.Push(&ri, &dest, plan, push.Insert, 2, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", false)

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

	err := push.Push(&ri, &dest, plan, push.Insert, 2, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", false)

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

	err := push.Push(&ri, &dest, plan, push.Insert, 2, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", false)

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

	err := push.Push(&ri, &dest, plan, push.Insert, 5, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", false)

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
	err := push.Push(ri, dest, plan, push.Insert, 10, 100*time.Millisecond, true, push.NoErrorCaptureRowWriter{}, nil, "", "", false)

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
	err = push.Push(&ri, dest, plan, push.Insert, 2, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", tmpfile.Name(), false)

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
	err := push.Push(&ri, dest, plan, push.Insert, 10, 0, true, push.NoErrorCaptureRowWriter{}, nil, "", "", false, obs)

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
