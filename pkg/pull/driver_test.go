package pull_test

import (
	"encoding/json"
	"testing"

	"github.com/cgi-fr/lino/pkg/pull"
	"github.com/stretchr/testify/assert"
)

func makeTable(name string) pull.Table {
	return pull.NewTable(name, []string{name + "_ID"}, nil)
}

func makeRel(from, to pull.Table) pull.Relation {
	return pull.NewRelation(from.Name()+to.Name(), from, to, []string{to.Name() + "_ID"}, []string{to.Name() + "_ID"})
}

func assertResult(t *testing.T, result []pull.ExportableRow, expected string) {
	b, err := json.Marshal(result)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(b))
}

func TestPull1(t *testing.T) {
	exporter := &MemoryRowExporter{rows: []pull.ExportableRow{}}

	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")

	AB := makeRel(A, B)
	BC := makeRel(B, C)

	step3 := pull.NewStep(3, C, BC, pull.NewRelationList([]pull.Relation{}), pull.NewCycleList([]pull.Cycle{}), pull.NewStepList([]pull.Step{}))
	step2 := pull.NewStep(2, B, AB, pull.NewRelationList([]pull.Relation{}), pull.NewCycleList([]pull.Cycle{}), pull.NewStepList([]pull.Step{step3}))
	step1 := pull.NewStep(1, A, nil, pull.NewRelationList([]pull.Relation{}), pull.NewCycleList([]pull.Cycle{}), pull.NewStepList([]pull.Step{step2}))

	plan := pull.NewPlan(
		pull.NewFilter(2, pull.Row{}, ""),
		pull.NewStepList([]pull.Step{step1, step2, step3}),
	)

	source := map[string][]pull.Row{
		A.Name(): {
			{A.PrimaryKey()[0]: 10, AB.ParentKey()[0]: 20},
			{A.PrimaryKey()[0]: 11, AB.ParentKey()[0]: 21},
			{A.PrimaryKey()[0]: 12, AB.ParentKey()[0]: 22},
		},
		B.Name(): {
			{B.PrimaryKey()[0]: 20, BC.ParentKey()[0]: 30},
			{B.PrimaryKey()[0]: 21, BC.ParentKey()[0]: 31},
			{B.PrimaryKey()[0]: 22, BC.ParentKey()[0]: 32},
		},
		C.Name(): {
			{C.PrimaryKey()[0]: 30},
			{C.PrimaryKey()[0]: 31},
			{C.PrimaryKey()[0]: 32},
		},
	}
	datasource := &MemoryDataSource{source}

	err := pull.Pull(plan, pull.NewOneEmptyRowReader(), datasource, exporter, pull.NoTraceListener{})

	assert.Nil(t, err)
	assert.Len(t, exporter.rows, int(plan.InitFilter().Limit()))

	assertResult(t, exporter.rows, `[{"A_ID":10,"B_ID":20,"AB":[{"B_ID":20,"C_ID":30,"BC":[{"C_ID":30}]}]},{"A_ID":11,"B_ID":21,"AB":[{"B_ID":21,"C_ID":31,"BC":[{"C_ID":31}]}]}]`)
}

func TestPull2(t *testing.T) {
	exporter := &MemoryRowExporter{rows: []pull.ExportableRow{}}

	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")

	AB := makeRel(A, B)
	AC := makeRel(A, C)

	step3 := pull.NewStep(3, C, AC, pull.NewRelationList([]pull.Relation{}), pull.NewCycleList([]pull.Cycle{}), pull.NewStepList([]pull.Step{}))
	step2 := pull.NewStep(2, B, AB, pull.NewRelationList([]pull.Relation{}), pull.NewCycleList([]pull.Cycle{}), pull.NewStepList([]pull.Step{}))
	step1 := pull.NewStep(1, A, nil, pull.NewRelationList([]pull.Relation{}), pull.NewCycleList([]pull.Cycle{}), pull.NewStepList([]pull.Step{step2, step3}))

	plan := pull.NewPlan(
		pull.NewFilter(2, pull.Row{}, ""),
		pull.NewStepList([]pull.Step{step1, step2, step3}),
	)

	source := map[string][]pull.Row{
		A.Name(): {
			{A.PrimaryKey()[0]: 10, AB.ParentKey()[0]: 20, AC.ParentKey()[0]: 30},
			{A.PrimaryKey()[0]: 11, AB.ParentKey()[0]: 21, AC.ParentKey()[0]: 31},
			{A.PrimaryKey()[0]: 12, AB.ParentKey()[0]: 22, AC.ParentKey()[0]: 32},
		},
		B.Name(): {
			{B.PrimaryKey()[0]: 20},
			{B.PrimaryKey()[0]: 21},
			{B.PrimaryKey()[0]: 22},
		},
		C.Name(): {
			{C.PrimaryKey()[0]: 30},
			{C.PrimaryKey()[0]: 31},
			{C.PrimaryKey()[0]: 32},
		},
	}
	datasource := &MemoryDataSource{source}

	err := pull.Pull(plan, pull.NewOneEmptyRowReader(), datasource, exporter, pull.NoTraceListener{})

	assert.Nil(t, err)
	assert.Len(t, exporter.rows, int(plan.InitFilter().Limit()))

	assertResult(t, exporter.rows, `[{"A_ID":10,"B_ID":20,"C_ID":30,"AB":[{"B_ID":20}],"AC":[{"C_ID":30}]},{"A_ID":11,"B_ID":21,"C_ID":31,"AB":[{"B_ID":21}],"AC":[{"C_ID":31}]}]`)
}

func TestPull3(t *testing.T) {
	exporter := &MemoryRowExporter{rows: []pull.ExportableRow{}}

	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")
	D := makeTable("D")

	AB := makeRel(A, B)
	AC := makeRel(A, C)
	BD := makeRel(B, D)
	CD := makeRel(C, D)

	step5 := pull.NewStep(5, D, CD, pull.NewRelationList([]pull.Relation{}), pull.NewCycleList([]pull.Cycle{}), pull.NewStepList([]pull.Step{}))
	step4 := pull.NewStep(4, D, BD, pull.NewRelationList([]pull.Relation{}), pull.NewCycleList([]pull.Cycle{}), pull.NewStepList([]pull.Step{}))
	step3 := pull.NewStep(3, C, AC, pull.NewRelationList([]pull.Relation{}), pull.NewCycleList([]pull.Cycle{}), pull.NewStepList([]pull.Step{step5}))
	step2 := pull.NewStep(2, B, AB, pull.NewRelationList([]pull.Relation{}), pull.NewCycleList([]pull.Cycle{}), pull.NewStepList([]pull.Step{step4}))
	step1 := pull.NewStep(1, A, nil, pull.NewRelationList([]pull.Relation{}), pull.NewCycleList([]pull.Cycle{}), pull.NewStepList([]pull.Step{step2, step3}))

	plan := pull.NewPlan(
		pull.NewFilter(2, pull.Row{}, ""),
		pull.NewStepList([]pull.Step{step1, step2, step3, step4, step5}),
	)

	source := map[string][]pull.Row{
		A.Name(): {
			{A.PrimaryKey()[0]: 10, AB.ParentKey()[0]: 20, AC.ParentKey()[0]: 30},
			{A.PrimaryKey()[0]: 11, AB.ParentKey()[0]: 21, AC.ParentKey()[0]: 31},
			{A.PrimaryKey()[0]: 12, AB.ParentKey()[0]: 22, AC.ParentKey()[0]: 32},
		},
		B.Name(): {
			{B.PrimaryKey()[0]: 20, BD.ParentKey()[0]: 40},
			{B.PrimaryKey()[0]: 21, BD.ParentKey()[0]: 41},
			{B.PrimaryKey()[0]: 22, BD.ParentKey()[0]: 42},
		},
		C.Name(): {
			{C.PrimaryKey()[0]: 30, CD.ParentKey()[0]: 40},
			{C.PrimaryKey()[0]: 31, CD.ParentKey()[0]: 41},
			{C.PrimaryKey()[0]: 32, CD.ParentKey()[0]: 42},
		},
		D.Name(): {
			{D.PrimaryKey()[0]: 40},
			{D.PrimaryKey()[0]: 41},
			{D.PrimaryKey()[0]: 42},
		},
	}
	datasource := &MemoryDataSource{source}

	err := pull.Pull(plan, pull.NewOneEmptyRowReader(), datasource, exporter, pull.NoTraceListener{})

	assert.Nil(t, err)
	assert.Len(t, exporter.rows, int(plan.InitFilter().Limit()))

	assertResult(t, exporter.rows, `[{"A_ID":10,"B_ID":20,"C_ID":30,"AB":[{"B_ID":20,"D_ID":40,"BD":[{"D_ID":40}]}],"AC":[{"C_ID":30,"D_ID":40,"CD":[{"D_ID":40}]}]},{"A_ID":11,"B_ID":21,"C_ID":31,"AB":[{"B_ID":21,"D_ID":41,"BD":[{"D_ID":41}]}],"AC":[{"C_ID":31,"D_ID":41,"CD":[{"D_ID":41}]}]}]`)
}

func TestPull4(t *testing.T) {
	exporter := &MemoryRowExporter{rows: []pull.ExportableRow{}}

	A := makeTable("A")
	B := makeTable("B")

	AB := makeRel(A, B)
	BA := makeRel(B, A)

	cycle1 := pull.NewRelationList([]pull.Relation{AB, BA})
	step1 := pull.NewStep(1, A, nil, cycle1, pull.NewCycleList([]pull.Cycle{cycle1}), pull.NewStepList([]pull.Step{}))

	plan := pull.NewPlan(
		pull.NewFilter(2, pull.Row{}, ""),
		pull.NewStepList([]pull.Step{step1}),
	)

	source := map[string][]pull.Row{
		A.Name(): {
			{A.PrimaryKey()[0]: 10, AB.ParentKey()[0]: 20},
			{A.PrimaryKey()[0]: 11, AB.ParentKey()[0]: 21},
			{A.PrimaryKey()[0]: 12, AB.ParentKey()[0]: 22},
		},
		B.Name(): {
			{B.PrimaryKey()[0]: 20, BA.ParentKey()[0]: 10},
			{B.PrimaryKey()[0]: 21, BA.ParentKey()[0]: 11},
			{B.PrimaryKey()[0]: 22, BA.ParentKey()[0]: 12},
		},
	}
	datasource := &MemoryDataSource{source}

	err := pull.Pull(plan, pull.NewOneEmptyRowReader(), datasource, exporter, pull.NoTraceListener{})

	assert.Nil(t, err)
	assert.Len(t, exporter.rows, int(plan.InitFilter().Limit()))

	assertResult(t, exporter.rows, `[{"A_ID":10,"B_ID":20,"AB":[{"A_ID":10,"B_ID":20}]},{"A_ID":11,"B_ID":21,"AB":[{"A_ID":11,"B_ID":21}]}]`)
}
