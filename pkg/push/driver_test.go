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
	"testing"

	"github.com/cgi-fr/lino/pkg/push"
	"github.com/stretchr/testify/assert"
)

func makeTable(name string) push.Table {
	return push.NewTable(name, []string{})
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
		A.Name(): &rowWriter{},
		B.Name(): &rowWriter{},
		C.Name(): &rowWriter{},
	}
	dest := memoryDataDestination{tables, false, false, false}

	err := push.Push(&ri, &dest, plan, push.Insert, 2, true, push.NoErrorCaptureRowWriter{})

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
		A.Name(): &rowWriter{},
		B.Name(): &rowWriter{},
		C.Name(): &rowWriter{},
	}
	dest := memoryDataDestination{tables, false, false, false}

	err := push.Push(&ri, &dest, plan, push.Insert, 2, true, push.NoErrorCaptureRowWriter{})

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
		A.Name(): &rowWriter{},
		B.Name(): &rowWriter{},
		C.Name(): &rowWriter{},
	}
	dest := memoryDataDestination{tables, false, false, false}

	err := push.Push(&ri, &dest, plan, push.Insert, 2, true, push.NoErrorCaptureRowWriter{})

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
		A.Name(): &rowWriter{},
		B.Name(): &rowWriter{},
		C.Name(): &rowWriter{},
	}
	dest := memoryDataDestination{tables, false, false, false}

	err := push.Push(&ri, &dest, plan, push.Insert, 5, true, push.NoErrorCaptureRowWriter{})

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
