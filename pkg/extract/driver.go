package extract

import "fmt"

var logger Logger = Nologger{}

// SetLogger if needed, default no logger
func SetLogger(l Logger) {
	logger = l
}

// Extract data from source following the given extraction plan.
func Extract(plan Plan, source DataSource, exporter RowExporter, diagnostic TraceListener) *Error {
	e := extractor{source}
	if err := e.extract(plan, exporter.Export, diagnostic); err != nil {
		return err
	}

	return nil
}

type extractor struct {
	datasource DataSource
}

func (e extractor) extract(plan Plan, export func(Row) *Error, diagnostic TraceListener) *Error {
	filter := plan.InitFilter()
	if err := e.extractStep(plan.Steps().Step(0), filter, export, diagnostic); err != nil {
		return err
	}
	return nil
}

func (e extractor) extractStep(step Step, filter Filter, export func(Row) *Error, diagnostic TraceListener) *Error {
	rows, err := e.read(step.Entry(), filter)
	diagnostic = diagnostic.TraceStep(step, filter)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("extract: from %v with filter %v returned %v row(s)", step.Entry(), filter, len(rows)))

	for i, row := range rows {
		logger.Trace(fmt.Sprintf("extract: process row number %v", i))

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
			logger.Trace(fmt.Sprintf("extract: row #%v, following %v from %v", i, rel, fromTable.Name()))
			relatedToRows := allRows[fromTable.Name()]
			logger.Trace(fmt.Sprintf("extract: row #%v, %v related row(s)", i, len(relatedToRows)))
			for _, relatedToRow := range relatedToRows {
				nextFilter := relatedTo(nextStep.Entry(), rel, relatedToRow, false)
				if relatedToRow[rel.Name()] == nil {
					relatedToRow[rel.Name()] = []Row{}
				}
				if err := e.extractStep(nextStep, nextFilter, func(r Row) *Error {
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

	return nil
}

func (e extractor) exhaust(step Step, allRows map[string][]Row) *Error {
	cycles := step.Cycles()

	logger.Trace(fmt.Sprintf("extract: %v cycle(s) to traverse", cycles.Len()))

	fromTable := step.Entry()
	for cycleIdx := uint(0); cycleIdx < step.Cycles().Len(); cycleIdx++ {
		cycle := step.Cycles().Cycle(cycleIdx)
		logger.Trace(fmt.Sprintf("extract: traversing cycle %v", cycle))
		for relationIdx := uint(0); relationIdx < cycle.Len(); relationIdx++ {
			relation := cycle.Relation(relationIdx)
			fromRows := allRows[fromTable.Name()]
			logger.Trace(fmt.Sprintf("extract: following relation %v has %v source row(s)", relation, len(fromRows)))
			for i, fromRow := range fromRows {
				logger.Trace(fmt.Sprintf("extract: following relation %v on row #%v (%v)", relation, i, fromRow))
				nextFilter := relatedTo(fromTable, relation, fromRow, true)
				logger.Trace(fmt.Sprintf("extract: following relation %v on row #%v with filter %v", relation, i, nextFilter))
				toTable := relation.OppositeOf(fromTable.Name())
				directionParent := toTable.Name() == relation.Parent().Name()
				rows, err := e.read(toTable, nextFilter)
				if err != nil {
					return err
				}

				logger.Trace(fmt.Sprintf("extract: following relation %v on row #%v returned %v related row(s)", relation, i, len(rows)))
				rows = removeDuplicate(toTable.PrimaryKey(), rows, allRows[toTable.Name()])
				logger.Trace(fmt.Sprintf("extract: following relation %v on row #%v returned %v unseen row(s)", relation, i, len(rows)))

				if len(rows) == 0 {
					logger.Trace(fmt.Sprintf("extract: stop traversing cycle %v", cycle))
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

func (e extractor) read(t Table, f Filter) ([]Row, *Error) {
	iter, err := e.datasource.Read(t, f)
	if err != nil {
		return nil, err
	}
	result := []Row{}
	for iter.Next() {
		row, err := iter.Value()
		if err != nil {
			return nil, err
		}
		result = append(result, row)
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

func relatedTo(from Table, follow Relation, data Row, exhaust bool) Filter {
	logger.Trace(fmt.Sprintf("extract: build filter with row %v and relation %v to extract data from table %v", data, follow, from))
	var row Row
	switch from.Name() {
	case follow.Parent().Name():
		if exhaust {
			logger.Trace(fmt.Sprintf("extract: build parent filter %v=data[%v]=%v", follow.ChildKey(), follow.ParentKey(), data[follow.ParentKey()]))
			row = Row{follow.ChildKey(): data[follow.ParentKey()]}
		} else {
			logger.Trace(fmt.Sprintf("extract: build parent filter %v=data[%v]=%v", follow.ParentKey(), follow.ChildKey(), data[follow.ChildKey()]))
			row = Row{follow.ParentKey(): data[follow.ChildKey()]}
		}
	case follow.Child().Name():
		logger.Trace(fmt.Sprintf("extract: build child filter %v=data[%v]=%v", follow.ParentKey(), follow.ChildKey(), data[follow.ChildKey()]))
		row = Row{follow.ParentKey(): data[follow.ChildKey()]}
	default:
		logger.Error(fmt.Sprintf("extract: cannot build filter with row %v and relation %v to extract data from table %v", data, follow, from))
		panic(nil)
	}

	return NewFilter(0, row)
}
func removeDuplicate(pk string, a, b []Row) []Row {
	result := []Row{}
loop:
	for _, row1 := range a {
		for _, row2 := range b {
			if row1[pk] == row2[pk] {
				continue loop
			}
		}
		result = append(result, row1)
	}
	return result
}
