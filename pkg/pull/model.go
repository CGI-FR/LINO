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

package pull

import (
	"encoding/json"

	over "github.com/Trendyol/overlog"
	"github.com/rs/zerolog/log"
)

// Table from which to pull data.
type Table interface {
	Name() string
	PrimaryKey() []string
	Columns() ColumnList
	export(Row) ExportableRow
}

// ColumnList is a list of columns.
type ColumnList interface {
	Len() uint
	Contains(string) bool
	Column(idx uint) Column

	add(Column) ColumnList
}

// Column of a table.
type Column interface {
	Name() string
	Export() string
}

// Relation between two tables.
type Relation interface {
	Name() string
	Parent() Table
	Child() Table
	ParentKey() []string
	ChildKey() []string
	OppositeOf(tablename string) Table
}

// RelationList is a list of relations.
type RelationList interface {
	Len() uint
	Relation(idx uint) Relation
}

// Cycle is a list of relations.
type Cycle interface {
	RelationList
}

// CycleList is a list of cycles.
type CycleList interface {
	Len() uint
	Cycle(idx uint) Cycle
}

// Step group of follows to perform.
type Step interface {
	Index() uint
	Entry() Table
	Follow() Relation
	Relations() RelationList
	Cycles() CycleList
	NextSteps() StepList
}

// StepList list of steps to perform.
type StepList interface {
	Len() uint
	Step(uint) Step
}

// Plan of the puller process.
type Plan interface {
	InitFilter() Filter
	Steps() StepList
}

// Filter applied to data tables.
type Filter interface {
	Limit() uint
	Values() Row
	Where() string
}

// Error is the error type returned by the domain
type Error struct {
	Description string
}

func (e *Error) Error() string {
	return e.Description
}

type ExecutionStats interface {
	GetLinesPerStepCount() map[string]int
	GetFiltersCount() int

	ToJSON() []byte
}

type stats struct {
	LinesPerStepCount map[string]int `json:"linesPerStepCount"`
	FiltersCount      int            `json:"filtersCount"`
}

// Reset all statistics to zero
func Reset() {
	over.MDC().Set("stats", &stats{LinesPerStepCount: map[string]int{}})
}

// Compute current statistics and give a snapshot
func Compute() ExecutionStats {
	value, exists := over.MDC().Get("stats")
	if stats, ok := value.(ExecutionStats); exists && ok {
		return stats
	}
	log.Warn().Msg("Unable to compute statistics")
	return &stats{}
}

// Exports statistics to readable json format
func (s *stats) ToJSON() []byte {
	b, err := json.Marshal(s)
	if err != nil {
		log.Warn().Msg("Unable to read statistics")
	}
	return b
}

func (s *stats) GetLinesPerStepCount() map[string]int {
	return s.LinesPerStepCount
}

func (s *stats) GetFiltersCount() int {
	return s.FiltersCount
}

func IncLinesPerStepCount(step string) {
	stats := getStats()
	stats.LinesPerStepCount[step]++
}

func IncFiltersCount() {
	stats := getStats()
	stats.FiltersCount++
}

// Compute current statistics and give a snapshot
func getStats() *stats {
	value, exists := over.MDC().Get("stats")
	if stats, ok := value.(*stats); exists && ok {
		return stats
	}
	log.Warn().Msg("Statistics uncorrectly initialized")
	return &stats{}
}
