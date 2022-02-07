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
	"fmt"

	"github.com/rs/zerolog/log"
)

type Step struct {
	p     *puller
	out   ExportedRow
	entry Relation
	cache DataSet
	next  []*Step
}

func (s *Step) Entry() Relation {
	return s.entry
}

func NewStep(puller *puller, out ExportedRow, entry Relation) *Step {
	return &Step{
		p:     puller,
		out:   out,
		entry: entry,
		cache: DataSet{},
		next:  []*Step{},
	}
}

type Puller interface {
	Pull(start Table, filter Filter, filterCohort RowReader) error
}

type puller struct {
	graph      Graph
	datasource DataSource
	exporter   RowExporter
	diagnostic TraceListener
}

func NewPuller(plan Plan, datasource DataSource, exporter RowExporter, diagnostic TraceListener) Puller {
	return &puller{
		graph:      plan.buildGraph(),
		datasource: datasource,
		exporter:   exporter,
		diagnostic: diagnostic,
	}
}

func (p *puller) Pull(start Table, filter Filter, filterCohort RowReader) error {
	start = p.graph.addMissingColumns(start)

	if err := p.datasource.Open(); err != nil {
		return fmt.Errorf("%w", err)
	}

	defer p.datasource.Close()

	Reset()

	filters := []Filter{}
	if filterCohort != nil {
		for filterCohort.Next() {
			fc := filterCohort.Value()
			values := Row{}
			for key, val := range fc {
				values[key] = val
			}
			for key, val := range filter.Values {
				values[key] = val
			}
			filters = append(filters, Filter{
				Limit:    filter.Limit,
				Values:   values,
				Where:    filter.Where,
				Distinct: filter.Distinct,
			})
		}
	} else {
		filters = append(filters, Filter{
			Limit:    filter.Limit,
			Values:   filter.Values,
			Where:    filter.Where,
			Distinct: filter.Distinct,
		})
	}

	for _, f := range filters {
		IncFiltersCount()
		reader, err := p.datasource.RowReader(start, f)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		for reader.Next() {
			IncLinesPerStepCount(string(start.Name))
			row := start.export(reader.Value())

			if err := p.pull(start, row); err != nil {
				return fmt.Errorf("%w", err)
			}

			if err := p.exporter.Export(row); err != nil {
				return fmt.Errorf("%w", err)
			}
		}

		if reader.Error() != nil {
			return fmt.Errorf("%w", reader.Error())
		}
	}

	return nil
}

func (p *puller) pull(source Table, out ExportedRow) error {
	relations, ok := p.graph.Relations[source.Name]
	if ok {
		for _, relation := range relations {
			if err := p.run(NewStep(p, out, relation)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *puller) run(step *Step) error {
	if err := step.Execute(); err != nil {
		return err
	}

	// This code seems to be counter-productive,
	// There may be a tipping point on len(step.next) where it become useful.
	// if len(step.next) > 10 {
	// 	group := errgroup.Group{}

	// 	for _, nextStep := range step.next {
	// 		step := nextStep

	// 		group.Go(func() error {
	// 			return p.run(step)
	// 		})
	// 	}

	// 	if err := group.Wait(); err != nil {
	// 		return fmt.Errorf("%w", err)
	// 	}
	// } else {
	for _, nextStep := range step.next {
		if err := p.run(nextStep); err != nil {
			return err
		}
	}
	// }

	return nil
}

func (s *Step) Execute() error {
	log.Trace().Interface("entry", s.entry.Name).Msg("begin step execution")
	s.p.diagnostic.TraceStep(*s)

	s.addToCache(s.entry.Local.Table, s.entry.Local.Table.getKeyValues(s.out))

	if err := s.follow(s.entry, s.out, s.p.graph.Components[s.entry.Foreign.Table.Name]); err != nil {
		return err
	}

	return nil
}

func (s *Step) pull(source Table, out ExportedRow, currentStep uint) error {
	relations, ok := s.p.graph.Relations[source.Name]
	if ok {
		for _, relation := range relations {
			if err := s.follow(relation, out, currentStep); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Step) follow(relation Relation, out ExportedRow, currentStep uint) error {
	if s.p.graph.Components[relation.Foreign.Table.Name] != currentStep {
		log.Trace().Interface("relation", relation.Name).Msg("edge of component reached")
		s.next = append(s.next, NewStep(s.p, out, relation))

		return nil
	}

	filter := createFilter(relation, out)

	rows, err := s.p.datasource.Read(relation.Foreign.Table, Filter{Limit: 0, Values: filter, Where: ""})
	IncFiltersCount()

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	exportedRows := s.removeDuplicates(relation.Foreign.Table, rows...)

	if relation.Cardinality == One {
		switch {
		case len(exportedRows) > 1:
			return fmt.Errorf("%w", ErrMultipleRowInOneToOneRelation)
		case len(exportedRows) == 1:
			out.Set(string(relation.Name), exportedRows[0])
		}
	} else {
		out.Set(string(relation.Name), exportedRows)
	}

	for _, row := range exportedRows {
		IncLinesPerStepCount(string(s.Entry().Name))
		if err := s.pull(relation.Foreign.Table, row, currentStep); err != nil {
			return err
		}
	}

	return nil
}

func (s *Step) removeDuplicates(table Table, rows ...Row) []ExportedRow {
	result := []ExportedRow{}

	// this table is not involved in a local tip (only foreign)
	// then it is not needed to cache seen pks
	if !s.p.graph.Cached[table.Name] {
		for _, row := range rows {
			result = append(result, table.export(row))
		}

		return result
	}

loop:
	for _, row1 := range rows {
		for _, row2 := range s.cache[table.Name] {
			all := true
			for _, pk := range table.Keys {
				all = all && row1[pk] == row2[pk]

				if !all {
					break
				}
			}
			if all {
				log.Trace().Interface("table", table.Name).Interface("row", row2).Msg("row removed because it has been seen")

				continue loop
			}
		}
		result = append(result, table.export(row1))
	}

	s.addToCache(table, rows...)

	return result
}

func (s *Step) addToCache(table Table, rows ...Row) {
	for _, row := range rows {
		seen := Row{}
		for _, pk := range table.Keys {
			seen[pk] = row[pk]
		}

		s.cache[table.Name] = append(s.cache[table.Name], seen)

		log.Trace().Interface("table", table.Name).Interface("seen", seen).Msg("update cache")
	}
}

func createFilter(relation Relation, localRow ExportedRow) map[string]interface{} {
	filter := Row{}

	for i := 0; i < len(relation.Foreign.Keys); i++ {
		foreignKey := relation.Foreign.Keys[i]
		localKey := relation.Local.Keys[i]
		filter[foreignKey] = localRow.GetOrNil(localKey)
	}

	return filter
}
