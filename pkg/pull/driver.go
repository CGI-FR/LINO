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

	e := puller{source}
	if err := e.pull(plan, filters, exporter.Export, diagnostic); err != nil {
		return err
	}

	return nil
}

type puller struct {
	datasource DataSource
}

func (e puller) pull(plan Plan, filters RowReader, export func(Row) *Error, diagnostic TraceListener) *Error {
	for filters.Next() {
		fileFilter := filters.Value()

		initFilter := filter{plan.InitFilter().Limit(), fileFilter.Update(plan.InitFilter().Values()), plan.InitFilter().Where()}
		if err := e.pullStep(plan.Steps().Step(0), initFilter, export, diagnostic); err != nil {
			return err
		}
	}

	return filters.Error()
}

func (e puller) pullStep(step Step, filter Filter, export func(Row) *Error, diagnostic TraceListener) *Error {
	rowIterator, err := e.datasource.RowReader(step.Entry(), filter)
	if err != nil {
		return err
	}
	diagnostic = diagnostic.TraceStep(step, filter)

	log.Info().Msg(fmt.Sprintf("pull: from %v with filter %v", step.Entry(), filter))

	i := 0
	for rowIterator.Next() {
		row := rowIterator.Value()
		i++
		log.Trace().Msg(fmt.Sprintf("pull: process row number %v", i))

		allRows := map[string][]Row{}
		allRows[step.Entry().Name()] = []Row{row}

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
			log.Trace().Msg(fmt.Sprintf("pull: row #%v, following %v from %v", i, rel, fromTable.Name()))
			relatedToRows := allRows[fromTable.Name()]
			log.Trace().Msg(fmt.Sprintf("pull: row #%v, %v related row(s)", i, len(relatedToRows)))
			for _, relatedToRow := range relatedToRows {
				nextFilter := relatedTo(nextStep.Entry(), rel, relatedToRow)
				if relatedToRow[rel.Name()] == nil {
					relatedToRow[rel.Name()] = []Row{}
				}
				if err := e.pullStep(nextStep, nextFilter, func(r Row) *Error {
					if !directionParent {
						rowArray, ok := relatedToRow[rel.Name()].([]Row)
						if !ok {
							return &Error{Description: fmt.Sprintf("table %v has a column whose name collides with the relation name %v", nextStep.Entry().Name(), rel.Name())}
						}
						rowArray = append(rowArray, r)
						relatedToRow[rel.Name()] = rowArray
					} else {
						relatedToRow[rel.Name()] = r
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

func (e puller) exhaust(step Step, allRows map[string][]Row) *Error {
	cycles := step.Cycles()

	log.Trace().Msg(fmt.Sprintf("pull: %v cycle(s) to traverse", cycles.Len()))

	fromTable := step.Entry()
	for cycleIdx := uint(0); cycleIdx < step.Cycles().Len(); cycleIdx++ {
		cycle := step.Cycles().Cycle(cycleIdx)
		log.Trace().Msg(fmt.Sprintf("pull: traversing cycle %v", cycle))
		for relationIdx := uint(0); relationIdx < cycle.Len(); relationIdx++ {
			relation := cycle.Relation(relationIdx)
			fromRows := allRows[fromTable.Name()]
			log.Trace().Msg(fmt.Sprintf("pull: following relation %v has %v source row(s)", relation, len(fromRows)))
			for i, fromRow := range fromRows {
				log.Trace().Msg(fmt.Sprintf("pull: following relation %v on row #%v (%v)", relation, i, fromRow))
				toTable := relation.OppositeOf(fromTable.Name())
				nextFilter := relatedTo(toTable, relation, fromRow)
				log.Trace().Msg(fmt.Sprintf("pull: following relation %v on row #%v with filter %v", relation, i, nextFilter))
				directionParent := toTable.Name() == relation.Parent().Name()
				rows, err := e.read(toTable, nextFilter)
				if err != nil {
					return err
				}

				log.Trace().Msg(fmt.Sprintf("pull: following relation %v on row #%v returned %v related row(s)", relation, i, len(rows)))
				rows = removeDuplicate(toTable.PrimaryKey(), rows, allRows[toTable.Name()])
				log.Trace().Msg(fmt.Sprintf("pull: following relation %v on row #%v returned %v unseen row(s)", relation, i, len(rows)))

				if len(rows) == 0 {
					log.Trace().Msg(fmt.Sprintf("pull: stop traversing cycle %v", cycle))
					break
				}

				if !directionParent {
					if fromRow[relation.Name()] == nil {
						fromRow[relation.Name()] = []Row{}
					}
					rowArray, ok := fromRow[relation.Name()].([]Row)
					if !ok {
						return &Error{Description: fmt.Sprintf("table %v has a column whose name collides with the relation name %v", fromTable.Name(), relation.Name())}
					}
					rowArray = append(rowArray, rows...)
					fromRow[relation.Name()] = rowArray
				} else {
					fromRow[relation.Name()] = rows[0]
				}

				allRows[toTable.Name()] = append(allRows[toTable.Name()], rows...)

				fromTable = toTable
			}
		}
	}

	return nil
}

func (e puller) read(t Table, f Filter) ([]Row, *Error) {
	iter, err := e.datasource.RowReader(t, f)
	if err != nil {
		return nil, err
	}
	result := []Row{}
	for iter.Next() {
		row := iter.Value()
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

func relatedTo(from Table, follow Relation, data Row) Filter {
	log.Trace().Msg(fmt.Sprintf("pull: build filter with row %v and relation %v to pull data from table %v", data, follow, from))
	if from.Name() != follow.Parent().Name() && from.Name() != follow.Child().Name() {
		log.Error().Msg(fmt.Sprintf("pull: cannot build filter with row %v and relation %v to pull data from table %v", data, follow, from))
		panic(nil)
	}

	if follow.Child().Name() == from.Name() {
		return NewFilter(0, buildFilterRow(follow.ChildKey(), follow.ParentKey(), data), "")
	}

	return NewFilter(0, buildFilterRow(follow.ParentKey(), follow.ChildKey(), data), "")
}

func removeDuplicate(pkList []string, a, b []Row) []Row {
	result := []Row{}
loop:
	for _, row1 := range a {
		for _, row2 := range b {
			all := true
			for _, pk := range pkList {
				all = all && row1[pk] == row2[pk]
			}
			if all {
				continue loop
			}
		}
		result = append(result, row1)
	}
	return result
}
