package extract_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"makeit.imfr.cgi.com/lino/pkg/extract"
)

var logger = Logger{}

func init() {
	extract.SetLogger(logger)
}

func makeTable(name string) extract.Table {
	return extract.NewTable(name, name+"_ID")
}

func makeRel(from, to extract.Table) extract.Relation {
	return extract.NewRelation(from.Name()+"->"+to.Name(), from, to, to.Name()+"_ID", to.Name()+"_ID")
}

/* func assertFollowedParent(t *testing.T, expected extract.Row, actual extract.Row, followed extract.Relation) []extract.Row {
	fmt.Printf("assert %v is equal to %v after following %v\n", actual, expected, &followed)
	cleanActual := extract.Row{}
	for key, value := range actual {
		if key != followed.Name && !strings.Contains(key, "->") {
			cleanActual[key] = value
		}
	}
	assert.Equal(t, expected, cleanActual)
	assert.NotNil(t, actual[followed.Name()])
	assert.IsType(t, []extract.Row{}, actual[followed.Name()])
	return actual[followed.Name()].([]extract.Row)
} */

func assertFollowedChild(t *testing.T, expected extract.Row, actual extract.Row, followed extract.Relation) []extract.Row {
	fmt.Printf("assert %v is equal to %v after following %v\n", actual, expected, &followed)
	cleanActual := extract.Row{}
	for key, value := range actual {
		if key != followed.Name() && !strings.Contains(key, "->") {
			cleanActual[key] = value
		}
	}
	assert.Equal(t, expected, cleanActual)
	assert.NotNil(t, actual[followed.Name()])
	assert.IsType(t, []extract.Row{}, actual[followed.Name()])
	return actual[followed.Name()].([]extract.Row)
}

func TestExtract1(t *testing.T) {
	exporter := &MemoryRowExporter{[]extract.Row{}}

	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")

	AB := makeRel(A, B)
	BC := makeRel(B, C)

	step3 := extract.NewStep(3, C, BC, extract.NewRelationList([]extract.Relation{}), extract.NewCycleList([]extract.Cycle{}), extract.NewStepList([]extract.Step{}))
	step2 := extract.NewStep(2, B, AB, extract.NewRelationList([]extract.Relation{}), extract.NewCycleList([]extract.Cycle{}), extract.NewStepList([]extract.Step{step3}))
	step1 := extract.NewStep(1, A, nil, extract.NewRelationList([]extract.Relation{}), extract.NewCycleList([]extract.Cycle{}), extract.NewStepList([]extract.Step{step2}))

	plan := extract.NewPlan(
		extract.NewFilter(2, extract.Row{}),
		extract.NewStepList([]extract.Step{step1, step2, step3}),
	)

	source := map[string][]extract.Row{
		A.Name(): {
			{A.PrimaryKey(): 10, AB.ParentKey(): 20},
			{A.PrimaryKey(): 11, AB.ParentKey(): 21},
			{A.PrimaryKey(): 12, AB.ParentKey(): 22},
		},
		B.Name(): {
			{B.PrimaryKey(): 20, BC.ParentKey(): 30},
			{B.PrimaryKey(): 21, BC.ParentKey(): 31},
			{B.PrimaryKey(): 22, BC.ParentKey(): 32},
		},
		C.Name(): {
			{C.PrimaryKey(): 30},
			{C.PrimaryKey(): 31},
			{C.PrimaryKey(): 32},
		},
	}
	datasource := &MemoryDataSource{source}

	err := extract.Extract(plan, datasource, exporter)

	assert.Nil(t, err)
	assert.Len(t, exporter.rows, int(plan.InitFilter().Limit()))

	B1 := assertFollowedChild(t, source[A.Name()][0], exporter.rows[0], AB)
	B2 := assertFollowedChild(t, source[A.Name()][1], exporter.rows[1], AB)

	C1 := assertFollowedChild(t, source[B.Name()][0], B1[0], BC)
	C2 := assertFollowedChild(t, source[B.Name()][1], B2[0], BC)

	assert.Equal(t, source[C.Name()][0], C1[0])
	assert.Equal(t, source[C.Name()][1], C2[0])
}

func TestExtract2(t *testing.T) {
	exporter := &MemoryRowExporter{[]extract.Row{}}

	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")

	AB := makeRel(A, B)
	AC := makeRel(A, C)

	step3 := extract.NewStep(3, C, AC, extract.NewRelationList([]extract.Relation{}), extract.NewCycleList([]extract.Cycle{}), extract.NewStepList([]extract.Step{}))
	step2 := extract.NewStep(2, B, AB, extract.NewRelationList([]extract.Relation{}), extract.NewCycleList([]extract.Cycle{}), extract.NewStepList([]extract.Step{}))
	step1 := extract.NewStep(1, A, nil, extract.NewRelationList([]extract.Relation{}), extract.NewCycleList([]extract.Cycle{}), extract.NewStepList([]extract.Step{step2, step3}))

	plan := extract.NewPlan(
		extract.NewFilter(2, extract.Row{}),
		extract.NewStepList([]extract.Step{step1, step2, step3}),
	)

	source := map[string][]extract.Row{
		A.Name(): {
			{A.PrimaryKey(): 10, AB.ParentKey(): 20, AC.ParentKey(): 30},
			{A.PrimaryKey(): 11, AB.ParentKey(): 21, AC.ParentKey(): 31},
			{A.PrimaryKey(): 12, AB.ParentKey(): 22, AC.ParentKey(): 32},
		},
		B.Name(): {
			{B.PrimaryKey(): 20},
			{B.PrimaryKey(): 21},
			{B.PrimaryKey(): 22},
		},
		C.Name(): {
			{C.PrimaryKey(): 30},
			{C.PrimaryKey(): 31},
			{C.PrimaryKey(): 32},
		},
	}
	datasource := &MemoryDataSource{source}

	err := extract.Extract(plan, datasource, exporter)

	assert.Nil(t, err)
	assert.Len(t, exporter.rows, int(plan.InitFilter().Limit()))

	B1 := assertFollowedChild(t, source[A.Name()][0], exporter.rows[0], AB)
	B2 := assertFollowedChild(t, source[A.Name()][1], exporter.rows[1], AB)

	assert.Equal(t, source[B.Name()][0], B1[0])
	assert.Equal(t, source[B.Name()][1], B2[0])

	C1 := assertFollowedChild(t, source[A.Name()][0], exporter.rows[0], AC)
	C2 := assertFollowedChild(t, source[A.Name()][1], exporter.rows[1], AC)

	assert.Equal(t, source[C.Name()][0], C1[0])
	assert.Equal(t, source[C.Name()][1], C2[0])
}

func TestExtract3(t *testing.T) {
	exporter := &MemoryRowExporter{[]extract.Row{}}

	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")
	D := makeTable("D")

	AB := makeRel(A, B)
	AC := makeRel(A, C)
	BD := makeRel(B, D)
	CD := makeRel(C, D)

	step5 := extract.NewStep(5, D, CD, extract.NewRelationList([]extract.Relation{}), extract.NewCycleList([]extract.Cycle{}), extract.NewStepList([]extract.Step{}))
	step4 := extract.NewStep(4, D, BD, extract.NewRelationList([]extract.Relation{}), extract.NewCycleList([]extract.Cycle{}), extract.NewStepList([]extract.Step{}))
	step3 := extract.NewStep(3, C, AC, extract.NewRelationList([]extract.Relation{}), extract.NewCycleList([]extract.Cycle{}), extract.NewStepList([]extract.Step{step5}))
	step2 := extract.NewStep(2, B, AB, extract.NewRelationList([]extract.Relation{}), extract.NewCycleList([]extract.Cycle{}), extract.NewStepList([]extract.Step{step4}))
	step1 := extract.NewStep(1, A, nil, extract.NewRelationList([]extract.Relation{}), extract.NewCycleList([]extract.Cycle{}), extract.NewStepList([]extract.Step{step2, step3}))

	plan := extract.NewPlan(
		extract.NewFilter(2, extract.Row{}),
		extract.NewStepList([]extract.Step{step1, step2, step3, step4, step5}),
	)

	source := map[string][]extract.Row{
		A.Name(): {
			{A.PrimaryKey(): 10, AB.ParentKey(): 20, AC.ParentKey(): 30},
			{A.PrimaryKey(): 11, AB.ParentKey(): 21, AC.ParentKey(): 31},
			{A.PrimaryKey(): 12, AB.ParentKey(): 22, AC.ParentKey(): 32},
		},
		B.Name(): {
			{B.PrimaryKey(): 20, BD.ParentKey(): 40},
			{B.PrimaryKey(): 21, BD.ParentKey(): 41},
			{B.PrimaryKey(): 22, BD.ParentKey(): 42},
		},
		C.Name(): {
			{C.PrimaryKey(): 30, CD.ParentKey(): 40},
			{C.PrimaryKey(): 31, CD.ParentKey(): 41},
			{C.PrimaryKey(): 32, CD.ParentKey(): 42},
		},
		D.Name(): {
			{D.PrimaryKey(): 40},
			{D.PrimaryKey(): 41},
			{D.PrimaryKey(): 42},
		},
	}
	datasource := &MemoryDataSource{source}

	err := extract.Extract(plan, datasource, exporter)

	assert.Nil(t, err)
	assert.Len(t, exporter.rows, int(plan.InitFilter().Limit()))

	B1 := assertFollowedChild(t, source[A.Name()][0], exporter.rows[0], AB)
	B2 := assertFollowedChild(t, source[A.Name()][1], exporter.rows[1], AB)

	C1 := assertFollowedChild(t, source[A.Name()][0], exporter.rows[0], AC)
	C2 := assertFollowedChild(t, source[A.Name()][1], exporter.rows[1], AC)

	D1 := assertFollowedChild(t, source[B.Name()][0], B1[0], BD)
	D2 := assertFollowedChild(t, source[B.Name()][1], B2[0], BD)
	D3 := assertFollowedChild(t, source[C.Name()][0], C1[0], CD)
	D4 := assertFollowedChild(t, source[C.Name()][1], C2[0], CD)

	assert.Equal(t, source[D.Name()][0], D1[0])
	assert.Equal(t, source[D.Name()][1], D2[0])
	assert.Equal(t, source[D.Name()][0], D3[0])
	assert.Equal(t, source[D.Name()][1], D4[0])
}

func TestExtract4(t *testing.T) {
	exporter := &MemoryRowExporter{[]extract.Row{}}

	A := makeTable("A")
	B := makeTable("B")

	AB := makeRel(A, B)
	BA := makeRel(B, A)

	cycle1 := extract.NewRelationList([]extract.Relation{AB, BA})
	step1 := extract.NewStep(1, A, nil, cycle1, extract.NewCycleList([]extract.Cycle{cycle1}), extract.NewStepList([]extract.Step{}))

	plan := extract.NewPlan(
		extract.NewFilter(2, extract.Row{}),
		extract.NewStepList([]extract.Step{step1}),
	)

	source := map[string][]extract.Row{
		A.Name(): {
			{A.PrimaryKey(): 10, AB.ParentKey(): 20},
			{A.PrimaryKey(): 11, AB.ParentKey(): 21},
			{A.PrimaryKey(): 12, AB.ParentKey(): 22},
		},
		B.Name(): {
			{B.PrimaryKey(): 20, BA.ParentKey(): 10},
			{B.PrimaryKey(): 21, BA.ParentKey(): 11},
			{B.PrimaryKey(): 22, BA.ParentKey(): 12},
		},
	}
	datasource := &MemoryDataSource{source}

	err := extract.Extract(plan, datasource, exporter)

	/* Expected result
	map[
		A_ID:10
		B_ID:20
		A->B:map[
			B_ID:20
			A_ID:10
		]
	] */

	assert.Nil(t, err)
	assert.Len(t, exporter.rows, int(plan.InitFilter().Limit()))

	B1 := assertFollowedChild(t, source[A.Name()][0], exporter.rows[0], AB)
	B2 := assertFollowedChild(t, source[A.Name()][1], exporter.rows[1], AB)

	assert.Equal(t, source[B.Name()][0], B1[0])
	assert.Equal(t, source[B.Name()][1], B2[0])
}
