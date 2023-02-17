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

	"github.com/cgi-fr/lino/pkg/push"
	"github.com/rs/zerolog/log"
)

type rowIterator struct {
	limit uint
	row   push.Row
}

func (ri *rowIterator) Error() *push.Error {
	return nil
}

func (ri *rowIterator) Value() *push.Row {
	return &ri.row
}

func (ri *rowIterator) Next() bool {
	if ri.limit == 0 {
		return false
	}
	ri.limit--

	return true
}

func (ri *rowIterator) Close() *push.Error {
	return nil
}

type memoryDataDestination struct {
	tables    map[string]*rowWriter
	closed    bool
	committed bool
	opened    bool
}

func (mdd *memoryDataDestination) RowWriter(table push.Table) (push.RowWriter, *push.Error) {
	return mdd.tables[table.Name()], nil
}

func (mdd *memoryDataDestination) Open(pla push.Plan, mode push.Mode, disableConstraints bool) *push.Error {
	mdd.opened = true
	return nil
}

func (mdd *memoryDataDestination) Commit() *push.Error {
	mdd.committed = true
	return nil
}

func (mdd *memoryDataDestination) Close() *push.Error {
	mdd.closed = true
	return nil
}

type rowWriter struct {
	rows []push.Row
}

func (rw *rowWriter) Write(row push.Row, where push.Row) *push.Error {
	log.Trace().Msg(fmt.Sprintf("append row %s to %s", row, rw.rows))
	rw.rows = append(rw.rows, row)
	return nil
}
