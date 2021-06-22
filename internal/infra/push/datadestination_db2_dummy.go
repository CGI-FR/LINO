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

// +build !db2

package push

import (
	"fmt"

	"github.com/cgi-fr/lino/pkg/push"
)

// Db2DataDestinationFactory exposes methods to create new Db2 extractors.
type Db2DataDestinationFactory struct{}

// NewDb2DataDestinationFactory creates a new Db2 datadestination factory.
func NewDb2DataDestinationFactory() *Db2DataDestinationFactory {
	return &Db2DataDestinationFactory{}
}

// New return a Db2 pusher
func (e *Db2DataDestinationFactory) New(url string, schema string) push.DataDestination {
	return NewSQLDataDestination(url, schema, Db2Dialect{})
}

// Db2Dialect inject oracle variations
type Db2Dialect struct{}

// Placeholde return the variable format for postgres
func (d Db2Dialect) Placeholder(position int) string {
	return "?"
}

// EnableConstraintsStatement generate statments to activate constraintes
func (d Db2Dialect) EnableConstraintsStatement(tableName string) string {
	panic(fmt.Errorf("Not implemented"))
}

// DisableConstraintsStatement generate statments to deactivate constraintes
func (d Db2Dialect) DisableConstraintsStatement(tableName string) string {
	panic(fmt.Errorf("Not implemented"))
}

// TruncateStatement generate statement to truncat table content
func (d Db2Dialect) TruncateStatement(tableName string) string {
	panic(fmt.Errorf("Not implemented"))
}

// InsertStatement generate insert statement
func (d Db2Dialect) InsertStatement(tableName string, columns []string, values []string, primaryKeys []string) string {
	panic(fmt.Errorf("Not implemented"))
}

// UpdateStatement
func (d Db2Dialect) UpdateStatement(tableName string, columns []string, uValues []string, primaryKeys []string, pValues []string) (string, *push.Error) {
	panic(fmt.Errorf("Not implemented"))
}

// IsDuplicateError check if error is a duplicate error
func (d Db2Dialect) IsDuplicateError(err error) bool {
	panic(fmt.Errorf("Not implemented"))
}

// ConvertValue before load
func (d Db2Dialect) ConvertValue(from push.Value) push.Value {
	panic(fmt.Errorf("Not implemented"))
}
