package load_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"makeit.imfr.cgi.com/lino/pkg/load"
)

var logger = Logger{}

func init() {
	load.SetLogger(logger)
}

func makeTable(name string) load.Table {
	return load.NewTable(name, name+"_ID")
}

func makeRel(from, to load.Table) load.Relation {
	return load.NewRelation(from.Name()+"->"+to.Name(), from, to)
}

func TestSimpleLoad(t *testing.T) {
	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")

	plan := load.NewPlan(
		A,
		[]load.Relation{},
	)
	ri := rowIterator{limit: 10, row: load.Row{"name": "John"}}
	tables := map[string]*rowWriter{
		A.Name(): &rowWriter{},
		B.Name(): &rowWriter{},
		C.Name(): &rowWriter{},
	}
	dest := memoryDataDestination{tables, false, false}

	err := load.Load(&ri, &dest, plan, load.Insert)

	assert.Nil(t, err)
	assert.Equal(t, true, dest.closed)
	assert.Equal(t, 10, len(dest.tables[A.Name()].rows))
	assert.Equal(t, "John", dest.tables[A.Name()].rows[0]["name"])

	assert.Equal(t, 0, len(dest.tables[B.Name()].rows))
}

func TestRelationLoad(t *testing.T) {
	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")

	AB := makeRel(A, B)
	BC := makeRel(B, C)

	plan := load.NewPlan(
		A,
		[]load.Relation{AB, BC},
	)
	ri := rowIterator{limit: 10, row: load.Row{
		"name": "John",
		"A->B": load.Row{
			"age": 42,
			"B->C": load.Row{
				"sex": "M",
			},
		}}}

	tables := map[string]*rowWriter{
		A.Name(): &rowWriter{},
		B.Name(): &rowWriter{},
		C.Name(): &rowWriter{},
	}
	dest := memoryDataDestination{tables, false, false}

	err := load.Load(&ri, &dest, plan, load.Insert)

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
