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

// Pull data from source following the given puller plan.
func Pull(plan Plan, filters RowReader, source DataSource, exporter RowExporter, diagnostic TraceListener) *Error {
	if err := source.Open(); err != nil {
		return err
	}

	defer source.Close()

	Reset()

	e := puller{source}
	if err := e.pull(plan, filters, exporter.Export, diagnostic); err != nil {
		return err
	}

	return nil
}

type puller struct {
	datasource DataSource
}

func (e puller) pull(plan Plan, filters RowReader, export func(ExportableRow) *Error, diagnostic TraceListener) *Error {
	for filters.Next() {
		IncFiltersCount()

		fileFilter := filters.Value()

		initFilter := filter{plan.InitFilter().Limit(), fileFilter.Update(plan.InitFilter().Values()), plan.InitFilter().Where(), plan.InitFilter().Distinct()}
		if err := e.pullStep(plan.Steps().Step(0), initFilter, export, diagnostic); err != nil {
			return err
		}
	}

	return filters.Error()
}

func (e puller) pullStep(step Step, filter Filter, export func(ExportableRow) *Error, diagnostic TraceListener) *Error {
	rowIterator, err := e.datasource.RowReader(step.Entry(), filter)
	if err != nil {
		return err
	}
	diagnostic = diagnostic.TraceStep(step, filter)

	log.Debug().Msg(fmt.Sprintf("from %v with filter %v", step.Entry(), filter))

	i := 0
	for rowIterator.Next() {
		row := step.Entry().export(rowIterator.Value())
		i++

		IncLinesPerStepCount(step.Entry().Name())
		log.Trace().Msg(fmt.Sprintf("process row number %v", i))

		allRows := map[string][]ExportableRow{}
		allRows[step.Entry().Name()] = []ExportableRow{row}

		if step.Relations().Len() > 0 {
			if err := e.exhaust(step, allRows); err != nil {
				return err
			}
		}

		for stepIdx := uint(0); stepIdx < step.NextSteps().Len(); stepIdx++ {
			nextStep := step.NextSteps().Step(stepIdx)
			rel := nextStep.Follow()
			fromTable := findFromTable(rel, step.Relations(), step.Entry())
			directionParent := rel.Child().Name() == fromTable.Name()
			log.Trace().Msg(fmt.Sprintf("row #%v, following %v from %v", i, rel, fromTable.Name()))
			relatedToRows := allRows[fromTable.Name()]
			log.Trace().Msg(fmt.Sprintf("row #%v, %v related row(s)", i, len(relatedToRows)))
			for _, relatedToRow := range relatedToRows {
				nextFilter := relatedTo(nextStep.Entry(), rel, relatedToRow)
				if err := e.pullStep(nextStep, nextFilter, func(r ExportableRow) *Error {
					if !directionParent {
						if !relatedToRow.add(rel.Name(), r) {
							return &Error{Description: fmt.Sprintf("table %v has a column whose name collides with the relation name %v", nextStep.Entry().Name(), rel.Name())}
						}
					} else {
						relatedToRow.set(rel.Name(), r)
					}
					return nil
				}, diagnostic); err != nil {
					return err
				}
			}
		}
		if err := export(row); err != nil {
			return err
		}
	}

	if rowIterator.Error() != nil {
		return rowIterator.Error()
	}
	return nil
}

func (e puller) exhaust(step Step, allRows map[string][]ExportableRow) *Error {
	cycles := step.Cycles()

	log.Trace().Msg(fmt.Sprintf("%v cycle(s) to traverse", cycles.Len()))

	fromTable := step.Entry()
	for cycleIdx := uint(0); cycleIdx < step.Cycles().Len(); cycleIdx++ {
		cycle := step.Cycles().Cycle(cycleIdx)
		log.Trace().Msg(fmt.Sprintf("traversing cycle %v", cycle))
		for relationIdx := uint(0); relationIdx < cycle.Len(); relationIdx++ {
			relation := cycle.Relation(relationIdx)
			fromRows := allRows[fromTable.Name()]
			log.Trace().Msg(fmt.Sprintf("following relation %v has %v source row(s)", relation, len(fromRows)))
			for i, fromRow := range fromRows {
				log.Trace().Msg(fmt.Sprintf("following relation %v on row #%v (%v)", relation, i, fromRow))
				toTable := relation.OppositeOf(fromTable.Name())
				nextFilter := relatedTo(toTable, relation, fromRow)
				log.Trace().Msg(fmt.Sprintf("following relation %v on row #%v with filter %v", relation, i, nextFilter))
				directionParent := toTable.Name() == relation.Parent().Name()
				rows, err := e.read(toTable, nextFilter)
				if err != nil {
					return err
				}

				log.Trace().Msg(fmt.Sprintf("following relation %v on row #%v returned %v related row(s)", relation, i, len(rows)))
				rows = removeDuplicate(toTable.PrimaryKey(), rows, allRows[toTable.Name()])
				log.Trace().Msg(fmt.Sprintf("following relation %v on row #%v returned %v unseen row(s)", relation, i, len(rows)))

				if len(rows) == 0 {
					log.Trace().Msg(fmt.Sprintf("stop traversing cycle %v", cycle))
					break
				}

				if !directionParent {
					if !fromRow.add(relation.Name(), rows...) {
						return &Error{Description: fmt.Sprintf("table %v has a column whose name collides with the relation name %v", fromTable.Name(), relation.Name())}
					}
				} else {
					fromRow.set(relation.Name(), rows[0])
				}

				allRows[toTable.Name()] = append(allRows[toTable.Name()], rows...)

				fromTable = toTable
			}
		}
	}

	return nil
}

func (e puller) read(t Table, f Filter) ([]ExportableRow, *Error) {
	iter, err := e.datasource.RowReader(t, f)
	if err != nil {
		return nil, err
	}
	result := []ExportableRow{}
	for iter.Next() {
		row := t.export(iter.Value())
		result = append(result, row)
	}
	if iter.Error() != nil {
		return nil, iter.Error()
	}
	return result, err
}

func findFromTable(rel Relation, relations RelationList, defaultTable Table) Table {
	for i := uint(0); i < relations.Len(); i++ {
		relation := relations.Relation(i)
		if rel.Child().Name() == relation.Parent().Name() {
			return rel.Child()
		}
		if rel.Parent().Name() == relation.Parent().Name() {
			return rel.Parent()
		}
		if rel.Child().Name() == relation.Child().Name() {
			return rel.Child()
		}
		if rel.Parent().Name() == relation.Child().Name() {
			return rel.Parent()
		}
	}
	return defaultTable
}

func buildFilterRow(targetKey []string, localKey []string, data Row) Row {
	row := Row{}
	for i := 0; i < len(targetKey); i++ {
		row[targetKey[i]] = data[localKey[i]]
	}
	return row
}

func relatedTo(from Table, follow Relation, data ExportableRow) Filter {
	log.Trace().Msg(fmt.Sprintf("build filter with row %v and relation %v to pull data from table %v", data, follow, from))
	if from.Name() != follow.Parent().Name() && from.Name() != follow.Child().Name() {
		log.Error().Msg(fmt.Sprintf("cannot build filter with row %v and relation %v to pull data from table %v", data, follow, from))
		panic(nil)
	}

	if follow.Child().Name() == from.Name() {
		return NewFilter(0, buildFilterRow(follow.ChildKey(), follow.ParentKey(), data.AsRow()), "", false)
	}

	return NewFilter(0, buildFilterRow(follow.ParentKey(), follow.ChildKey(), data.AsRow()), "", false)
}

func removeDuplicate(pkList []string, a, b []ExportableRow) []ExportableRow {
	result := []ExportableRow{}
loop:
	for _, row1 := range a {
		for _, row2 := range b {
			all := true
			for _, pk := range pkList {
				all = all && row1.GetOrNil(pk) == row2.GetOrNil(pk)
			}
			if all {
				continue loop
			}
		}
		result = append(result, row1)
	}
	return result
}
