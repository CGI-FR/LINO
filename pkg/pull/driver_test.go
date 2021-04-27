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

package pull_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cgi-fr/lino/pkg/pull"
	"github.com/stretchr/testify/assert"
)

var logger = Logger{}

func init() {
	pull.SetLogger(logger)
}

func makeTable(name string) pull.Table {
	return pull.NewTable(name, []string{name + "_ID"})
}

func makeRel(from, to pull.Table) pull.Relation {
	return pull.NewRelation(from.Name()+"->"+to.Name(), from, to, []string{to.Name() + "_ID"}, []string{to.Name() + "_ID"})
}

/* func assertFollowedParent(t *testing.T, expected pull.Row, actual pull.Row, followed pull.Relation) []pull.Row {
	fmt.Printf("assert %v is equal to %v after following %v\n", actual, expected, &followed)
	cleanActual := pull.Row{}
	for key, value := range actual {
		if key != followed.Name && !strings.Contains(key, "->") {
			cleanActual[key] = value
		}
	}
	assert.Equal(t, expected, cleanActual)
	assert.NotNil(t, actual[followed.Name()])
	assert.IsType(t, []pull.Row{}, actual[followed.Name()])
	return actual[followed.Name()].([]pull.Row)
} */

func assertFollowedChild(t *testing.T, expected pull.Row, actual pull.Row, followed pull.Relation) []pull.Row {
	fmt.Printf("assert %v is equal to %v after following %v\n", actual, expected, followed)
	cleanActual := pull.Row{}
	for key, value := range actual {
		if key != followed.Name() && !strings.Contains(key, "->") {
			cleanActual[key] = value
		}
	}
	assert.Equal(t, expected, cleanActual)
	assert.NotNil(t, actual[followed.Name()])
	assert.IsType(t, []pull.Row{}, actual[followed.Name()])
	return actual[followed.Name()].([]pull.Row)
}

func TestPull1(t *testing.T) {
	exporter := &MemoryRowExporter{[]pull.Row{}}

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

	B1 := assertFollowedChild(t, source[A.Name()][0], exporter.rows[0], AB)
	B2 := assertFollowedChild(t, source[A.Name()][1], exporter.rows[1], AB)

	C1 := assertFollowedChild(t, source[B.Name()][0], B1[0], BC)
	C2 := assertFollowedChild(t, source[B.Name()][1], B2[0], BC)

	assert.Equal(t, source[C.Name()][0], C1[0])
	assert.Equal(t, source[C.Name()][1], C2[0])
}

func TestPull2(t *testing.T) {
	exporter := &MemoryRowExporter{[]pull.Row{}}

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

	B1 := assertFollowedChild(t, source[A.Name()][0], exporter.rows[0], AB)
	B2 := assertFollowedChild(t, source[A.Name()][1], exporter.rows[1], AB)

	assert.Equal(t, source[B.Name()][0], B1[0])
	assert.Equal(t, source[B.Name()][1], B2[0])

	C1 := assertFollowedChild(t, source[A.Name()][0], exporter.rows[0], AC)
	C2 := assertFollowedChild(t, source[A.Name()][1], exporter.rows[1], AC)

	assert.Equal(t, source[C.Name()][0], C1[0])
	assert.Equal(t, source[C.Name()][1], C2[0])
}

func TestPull3(t *testing.T) {
	exporter := &MemoryRowExporter{[]pull.Row{}}

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

func TestPull4(t *testing.T) {
	exporter := &MemoryRowExporter{[]pull.Row{}}

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
