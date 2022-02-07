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

func TestOrderTableShouldRespectChildrenFirstParentsLast(t *testing.T) {
	t.Parallel()

	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")

	AB := makeRel(A, B)
	BC := makeRel(B, C)

	plan := push.NewPlan(
		A,
		[]push.Relation{AB, BC},
	)

	tables := plan.Tables()

	assert.Equal(t, []push.Table{C, B, A}, tables)
}

func TestOrderTableShouldRespectChildrenFirstParentsLast2(t *testing.T) {
	t.Parallel()

	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")
	D := makeTable("D")

	AB := makeRel(A, B)
	BC := makeRel(B, C)
	AD := makeRel(A, D)

	plan := push.NewPlan(
		A,
		[]push.Relation{BC, AB, AD},
	)

	tables := plan.Tables()

	assert.Contains(t, [][]push.Table{{C, B, D, A}, {C, D, B, A}}, tables)
}

func TestOrderTableShouldRespectChildrenFirstParentsLast3(t *testing.T) {
	t.Parallel()

	A := makeTable("A")
	B := makeTable("B")
	C := makeTable("C")
	D := makeTable("D")

	AB := makeRel(A, B)
	BC := makeRel(B, C)
	AD := makeRel(A, D)
	CD := makeRel(C, D)

	plan := push.NewPlan(
		A,
		[]push.Relation{BC, AB, AD, CD},
	)

	tables := plan.Tables()

	assert.Equal(t, []push.Table{D, C, B, A}, tables)
}
