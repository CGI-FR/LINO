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

package id_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cgi-fr/lino/pkg/id"
	"github.com/stretchr/testify/assert"
)

// MemoryStorage allows to store ingress descriptor objects in memory.
type MemoryStorage struct {
	id id.IngressDescriptor
}

func (s *MemoryStorage) Store(id id.IngressDescriptor) *id.Error {
	s.id = id
	return nil
}

func (s *MemoryStorage) Read() (id.IngressDescriptor, *id.Error) {
	return s.id, nil
}

// MockRelationReader read relations.
type MockRelationReader struct {
	fn func() (id.RelationList, *id.Error)
}

func (r *MockRelationReader) Read() (id.RelationList, *id.Error) {
	return r.fn()
}

// relation help to create id.Relation object from a string representation `parent -> child`.
func relationString(relation string) id.Relation {
	parentChild := strings.SplitN(relation, "->", 2)
	if parentChild == nil {
		return nil
	}

	parent := strings.TrimSpace(parentChild[0])
	child := strings.TrimSpace(parentChild[1])
	name := parent + "_" + child

	return id.NewRelation(name, id.NewTable(parent), id.NewTable(child))
}

// relation help to create id.Relation object from a string representation `parent -> child`.
func adRelationString(relation string, lookupParent bool, lookupChild bool) id.IngressRelation {
	return id.NewIngressRelation(relationString(relation), lookupParent, lookupChild, "", "", []string{}, []string{})
}

var adCreateTests = []struct {
	startTable string
	relations  id.RelationList
	id         id.IngressDescriptor
}{
	{
		"A",
		id.NewRelationList([]id.Relation{
			relationString("B->A"),
		}),
		id.NewIngressDescriptor(
			id.NewTable("A"),
			[]string{},
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("B->A", false, false),
			}),
		),
	},
	// more complex relations
	{
		"A",
		id.NewRelationList([]id.Relation{
			relationString("A->B"),
			relationString("B->C"),
			relationString("C->D"),
			relationString("Z->A"),
			relationString("Y->D"),
			relationString("Y->X"),
		}),
		id.NewIngressDescriptor(
			id.NewTable("A"),
			[]string{},
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("A->B", false, true),
				adRelationString("B->C", false, true),
				adRelationString("C->D", false, true),
				adRelationString("Z->A", false, false),
				adRelationString("Y->D", false, false),
				adRelationString("Y->X", false, false),
			}),
		),
	},
	// unrelated relations
	{
		"A",
		id.NewRelationList([]id.Relation{
			relationString("A->B"),
			relationString("B->C"),
			relationString("C->D"),
			relationString("Z->E"),
			relationString("Y->Z"),
			relationString("Y->X"),
		}),
		id.NewIngressDescriptor(
			id.NewTable("A"),
			[]string{},
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("A->B", false, true),
				adRelationString("B->C", false, true),
				adRelationString("C->D", false, true),
			}),
		),
	},
}

func TestCreate(t *testing.T) {
	for i, tt := range adCreateTests {
		t.Run(fmt.Sprintf("test create id %d", i), func(t *testing.T) {
			startTable := tt.startTable
			relReader := &MockRelationReader{
				fn: func() (id.RelationList, *id.Error) {
					return tt.relations, nil
				},
			}
			storage := &MemoryStorage{}

			err := id.Create(startTable, []string{}, relReader, storage)

			assert.Nil(t, err)

			assert.Equal(t, tt.id, storage.id)
		})
	}
}

func newInitialStep(tableName string) id.Step {
	table := id.NewTable(tableName)
	return id.NewStep(
		1,
		table,
		id.NewIngressRelation(id.NewRelation("", nil, nil), false, false, "", "", []string{}, []string{}),
		id.NewIngressRelationList([]id.IngressRelation{}),
		id.NewTableList([]id.Table{table}),
		id.NewCycleList([]id.IngressRelationList{}),
		0,
	)
}

func newSimpleStep(index uint, tableName string, relation string, lookupParent bool, lookupChild bool, from uint) id.Step {
	table := id.NewTable(tableName)
	return id.NewStep(
		index,
		table,
		adRelationString(relation, lookupParent, lookupChild),
		id.NewIngressRelationList([]id.IngressRelation{}),
		id.NewTableList([]id.Table{table}),
		id.NewCycleList([]id.IngressRelationList{}),
		from,
	)
}

/* func newLoopStep(index uint, tableName string, following string, tablenames []string, cycles []id.IngressRelationList, lookupParent bool, lookupChild bool, from uint) id.Step {
	tables := []id.Table{}
	for _, tabname := range tablenames {
		tables = append(tables, id.NewTable(tabname))
	}

	table := id.NewTable(tableName)
	return id.NewStep(
		index,
		table,
		adRelationString(following, lookupParent, lookupChild),
		id.NewTableList(tables),
		id.NewCycleList(cycles),
		from,
	)
} */

var adShowTests = []struct {
	id    id.IngressDescriptor
	steps []id.Step
}{
	{
		id.NewIngressDescriptor(
			id.NewTable("A"),
			[]string{},
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("A->B", false, true),
			}),
		),
		[]id.Step{
			newInitialStep("A"),
			newSimpleStep(2, "B", "A->B", false, true, 1),
		},
	},

	{ // example 1
		id.NewIngressDescriptor(
			id.NewTable("I"),
			[]string{},
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("C->O", false, true),
				adRelationString("O->D", false, true),
				adRelationString("I->D", false, true),
			}),
		),
		[]id.Step{
			newInitialStep("I"),
			newSimpleStep(2, "D", "I->D", false, true, 1),
		},
	},
	{ // example 1
		id.NewIngressDescriptor(
			id.NewTable("I"),
			[]string{},
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("C->O", false, true),
				adRelationString("O->D", false, true),
				adRelationString("I->D", false, true),
			}),
		),
		[]id.Step{
			newInitialStep("I"),
			newSimpleStep(2, "D", "I->D", false, true, 1),
		},
	},
	{ // example 1 (table C)
		id.NewIngressDescriptor(
			id.NewTable("C"),
			[]string{},
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("C->O", false, true),
				adRelationString("O->D", false, true),
				adRelationString("I->D", false, true),
			}),
		),
		[]id.Step{
			newInitialStep("C"),
			newSimpleStep(2, "O", "C->O", false, true, 1),
			newSimpleStep(3, "D", "O->D", false, true, 2),
		},
	},
	{ // example 1 (table D)
		id.NewIngressDescriptor(
			id.NewTable("D"),
			[]string{},
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("C->O", false, true),
				adRelationString("O->D", false, true),
				adRelationString("I->D", false, true),
			}),
		),
		[]id.Step{
			newInitialStep("D"),
		},
	},
	{ // example 1 (table O)
		id.NewIngressDescriptor(
			id.NewTable("O"),
			[]string{},
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("C->O", false, true),
				adRelationString("O->D", false, true),
				adRelationString("I->D", false, true),
			}),
		),
		[]id.Step{
			newInitialStep("O"),
			newSimpleStep(2, "D", "O->D", false, true, 1),
		},
	},
	{ // example 2
		id.NewIngressDescriptor(
			id.NewTable("C"),
			[]string{},
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("C->O", false, true),
				adRelationString("O->D", false, true),
				adRelationString("I->D", true, false),
			}),
		),
		[]id.Step{
			newInitialStep("C"),
			newSimpleStep(2, "O", "C->O", false, true, 1),
			newSimpleStep(3, "D", "O->D", false, true, 2),
			newSimpleStep(4, "I", "I->D", true, false, 3),
		},
	},
	/* { // example 3 : OK but order of tables change at each execution, need to fix it
		id.NewIngressDescriptor(
			id.NewTable("O"),
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("C->O", true, true),
			}),
		),
		[]id.Step{
			id.NewStep(
				1,
				id.NewTable("O"),
				id.NewIngressRelation(id.NewRelation("", nil, nil), false, false),
				id.NewIngressRelationList([]id.IngressRelation{}),
				id.NewTableList([]id.Table{id.NewTable("O"), id.NewTable("C")}),
				id.NewCycleList([]id.IngressRelationList{
					id.NewIngressRelationList([]id.IngressRelation{
						adRelationString("C->O", true, true),
					}),
				}),
				0,
			),
		},
	}, */
	/* { // example 3 Variant : OK but order of tables change at each execution, need to fix it
		id.NewIngressDescriptor(
			id.NewTable("D"),
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("D->O", false, true),
				adRelationString("C->O", true, true),
			}),
		),
		[]id.Step{
			newInitialStep("D"),
			newLoopStep(2, "O", "D->O", []string{"C", "O"}, []id.IngressRelationList{id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("C->O", true, true),
			})}, false, true, 1),
		},
	}, */
	{ // example 3 Variant bis
		id.NewIngressDescriptor(
			id.NewTable("O"),
			[]string{},
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("C->O", true, false),
			}),
		),
		[]id.Step{
			newInitialStep("O"),
			newSimpleStep(2, "C", "C->O", true, false, 1),
		},
	},
	/* { // example 4 FAIL
		id.NewIngressDescriptor(
			id.NewTable("I"),
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("C->O", true, false),
				adRelationString("O->D", true, true),
				adRelationString("I->D", true, true),
			}),
		),
		[]id.Step{
			newInitialStep("I"),
			newSimpleStep(2, "D", "I->D", true, false, 1),
			newSimpleStep(3, "O", "O->D", true, true, 2),
			newSimpleStep(4, "C", "C->O", true, false, 3),
			newSimpleStep(5, "D", "O->D", true, true, 3),
			newSimpleStep(6, "I", "I->D", true, false, 5),
		},
	}, */
	/* { // LOOP : OK but order of tables change at each execution, need to fix it
		id.NewIngressDescriptor(
			id.NewTable("O"),
			id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("O->D", false, true),
				adRelationString("I->D", false, true),
				adRelationString("I->E", true, false),
				adRelationString("D->E", false, true),
			}),
		),
		[]id.Step{
			newInitialStep("O"),
			newLoopStep(2, "D", "O->D", []string{"E", "D", "I"}, []id.IngressRelationList{id.NewIngressRelationList([]id.IngressRelation{
				adRelationString("D->E", false, true),
				adRelationString("I->E", true, false),
				adRelationString("I->D", false, true),
			})}, false, true, 1),
		},
	}, */
}

func TestUpdateStartTable(t *testing.T) {
	storage := &MemoryStorage{id: id.NewIngressDescriptor(id.NewTable("old"), []string{}, id.NewIngressRelationList([]id.IngressRelation{
		adRelationString("old->new", false, true),
	}))}

	err := id.SetStartTable(id.NewTable("new"), storage)

	assert.Nil(t, err)
	assert.Equal(t, id.NewIngressDescriptor(id.NewTable("new"), []string{}, id.NewIngressRelationList([]id.IngressRelation{
		adRelationString("old->new", false, true),
	})), storage.id)
}

func TestUpdateStartTableCheckTable(t *testing.T) {
	storage := &MemoryStorage{id: id.NewIngressDescriptor(id.NewTable("old"), []string{}, id.NewIngressRelationList([]id.IngressRelation{}))}

	err := id.SetStartTable(id.NewTable("new"), storage)

	assert.EqualError(t, err, "Table new doesn't exist")
	assert.Equal(t, id.NewIngressDescriptor(id.NewTable("old"), []string{}, id.NewIngressRelationList([]id.IngressRelation{})), storage.id)
}

func TestUpdateParentLookup(t *testing.T) {
	storage := &MemoryStorage{id: id.NewIngressDescriptor(id.NewTable("A"), []string{}, id.NewIngressRelationList([]id.IngressRelation{
		adRelationString("A->B", false, true),
	}))}

	err := id.SetParentLookup("A_B", true, storage)

	assert.Nil(t, err)
	assert.Equal(t, id.NewIngressDescriptor(id.NewTable("A"), []string{}, id.NewIngressRelationList([]id.IngressRelation{
		adRelationString("A->B", true, true),
	})), storage.id)
}

func TestUpdateChildLookup(t *testing.T) {
	storage := &MemoryStorage{id: id.NewIngressDescriptor(id.NewTable("A"), []string{}, id.NewIngressRelationList([]id.IngressRelation{
		adRelationString("A->B", false, true),
	}))}

	err := id.SetChildLookup("A_B", false, storage)

	assert.Nil(t, err)
	assert.Equal(t, id.NewIngressDescriptor(id.NewTable("A"), []string{}, id.NewIngressRelationList([]id.IngressRelation{
		adRelationString("A->B", false, false),
	})), storage.id)
}

func TestGetSteps(t *testing.T) {
	for i, tt := range adShowTests {
		t.Run(fmt.Sprintf("test get step %d", i), func(t *testing.T) {
			storage := &MemoryStorage{id: tt.id}

			ep, err := id.GetPullerPlan(storage)

			assert.Nil(t, err)

			for i := uint(0); i < ep.Len(); i++ {
				expectedStep := tt.steps[i]
				actualStep := ep.Step(i)
				assert.Equal(t, expectedStep, actualStep)
			}
		})
	}
}
