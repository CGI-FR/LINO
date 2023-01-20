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
	"time"

	over "github.com/adrienaury/zeromdc"
	"github.com/cgi-fr/jsonline/pkg/jsonline"
	"github.com/rs/zerolog/log"
)

type Cardinality bool

const (
	Many Cardinality = true
	One  Cardinality = false
)

type (
	TableName    string
	RelationName string
)

type Column struct {
	Name   string
	Export string
}

type ExportMode byte

const (
	ExportModeOnly ExportMode = iota
	ExportModeAll
)

type Table struct {
	Name       TableName
	Keys       []string
	Columns    []Column
	ExportMode ExportMode

	template jsonline.Template
}

type RelationTip struct {
	Table Table
	Keys  []string
}

type Relation struct {
	Name        RelationName
	Cardinality Cardinality
	Local       RelationTip
	Foreign     RelationTip
}

type RelationSet []Relation

type Plan struct {
	Relations  RelationSet
	Components map[TableName]uint // <= could be deduced from relations with tarjan algorithm
}

type Graph struct {
	Relations  map[TableName]RelationSet
	Components map[TableName]uint
	Cached     map[TableName]bool
}

type Row map[string]interface{}

type RowSet []Row

type DataSet map[TableName]RowSet

type Filter struct {
	Limit    uint
	Values   Row
	Where    string
	Distinct bool
}

// ExportedRow is a row but with keys ordered and values in export format for jsonline.
type ExportedRow struct {
	jsonline.Row
}

func (er ExportedRow) GetOrNil(key string) interface{} {
	v, _ := er.Get(key)

	return v
}

type ExecutionStats interface {
	GetLinesPerStepCount() map[string]int
	GetFiltersCount() int
	GetDuration() time.Duration

	ToJSON() []byte
}

type stats struct {
	LinesPerStepCount map[string]int `json:"linesPerStepCount"`
	FiltersCount      int            `json:"filtersCount"`
	Duration          time.Duration  `json:"duration"`
}

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

func (s *stats) GetDuration() time.Duration {
	return s.Duration
}

func IncLinesPerStepCount(step string) {
	stats := getStats()
	stats.LinesPerStepCount[step]++
}

func IncFiltersCount() {
	stats := getStats()
	stats.FiltersCount++
}

func SetDuration(duration time.Duration) {
	stats := getStats()
	stats.Duration = duration
}

func getStats() *stats {
	value, exists := over.MDC().Get("stats")
	if stats, ok := value.(*stats); exists && ok {
		return stats
	}
	log.Warn().Msg("Statistics uncorrectly initialized")
	return &stats{
		LinesPerStepCount: map[string]int{},
		FiltersCount:      0,
	}
}

func Compute() ExecutionStats {
	value, exists := over.MDC().Get("stats")
	if stats, ok := value.(ExecutionStats); exists && ok {
		return stats
	}
	log.Warn().Msg("Unable to compute statistics")
	return &stats{}
}

func Reset() {
	over.MDC().Set("stats", &stats{FiltersCount: 0, LinesPerStepCount: map[string]int{}})
}
