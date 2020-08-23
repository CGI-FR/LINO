package pull

import "fmt"

var logger Logger = Nologger{}

// SetLogger if needed, default no logger
func SetLogger(l Logger) {
	logger = l
}

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
		fileFilter, err := filters.Value()
		if err != nil {
			return err
		}
		initFilter := filter{plan.InitFilter().Limit(), fileFilter.Update(plan.InitFilter().Values())}
		if err := e.pullStep(plan.Steps().Step(0), initFilter, export, diagnostic); err != nil {
			return err
		}
	}
	return nil
}

func (e puller) pullStep(step Step, filter Filter, export func(Row) *Error, diagnostic TraceListener) *Error {
	rows, err := e.read(step.Entry(), filter)
	diagnostic = diagnostic.TraceStep(step, filter)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("pull: from %v with filter %v returned %v row(s)", step.Entry(), filter, len(rows)))

	for i, row := range rows {
		logger.Trace(fmt.Sprintf("pull: process row number %v", i))

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
			logger.Trace(fmt.Sprintf("pull: row #%v, following %v from %v", i, rel, fromTable.Name()))
			relatedToRows := allRows[fromTable.Name()]
			logger.Trace(fmt.Sprintf("pull: row #%v, %v related row(s)", i, len(relatedToRows)))
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

	return nil
}

func (e puller) exhaust(step Step, allRows map[string][]Row) *Error {
	cycles := step.Cycles()

	logger.Trace(fmt.Sprintf("pull: %v cycle(s) to traverse", cycles.Len()))

	fromTable := step.Entry()
	for cycleIdx := uint(0); cycleIdx < step.Cycles().Len(); cycleIdx++ {
		cycle := step.Cycles().Cycle(cycleIdx)
		logger.Trace(fmt.Sprintf("pull: traversing cycle %v", cycle))
		for relationIdx := uint(0); relationIdx < cycle.Len(); relationIdx++ {
			relation := cycle.Relation(relationIdx)
			fromRows := allRows[fromTable.Name()]
			logger.Trace(fmt.Sprintf("pull: following relation %v has %v source row(s)", relation, len(fromRows)))
			for i, fromRow := range fromRows {
				logger.Trace(fmt.Sprintf("pull: following relation %v on row #%v (%v)", relation, i, fromRow))
				toTable := relation.OppositeOf(fromTable.Name())
				nextFilter := relatedTo(toTable, relation, fromRow)
				logger.Trace(fmt.Sprintf("pull: following relation %v on row #%v with filter %v", relation, i, nextFilter))
				directionParent := toTable.Name() == relation.Parent().Name()
				rows, err := e.read(toTable, nextFilter)
				if err != nil {
					return err
				}

				logger.Trace(fmt.Sprintf("pull: following relation %v on row #%v returned %v related row(s)", relation, i, len(rows)))
				rows = removeDuplicate(toTable.PrimaryKey(), rows, allRows[toTable.Name()])
				logger.Trace(fmt.Sprintf("pull: following relation %v on row #%v returned %v unseen row(s)", relation, i, len(rows)))

				if len(rows) == 0 {
					logger.Trace(fmt.Sprintf("pull: stop traversing cycle %v", cycle))
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

func buildFilterRow(targetKey []string, localKey []string, data Row) Row {
	row := Row{}
	for i := 0; i < len(targetKey); i++ {
		row[targetKey[i]] = data[localKey[i]]
	}
	return row
}

func relatedTo(from Table, follow Relation, data Row) Filter {
	logger.Trace(fmt.Sprintf("pull: build filter with row %v and relation %v to pull data from table %v", data, follow, from))
	if from.Name() != follow.Parent().Name() && from.Name() != follow.Child().Name() {
		logger.Error(fmt.Sprintf("pull: cannot build filter with row %v and relation %v to pull data from table %v", data, follow, from))
		panic(nil)
	}

	if follow.Child().Name() == from.Name() {
		return NewFilter(0, buildFilterRow(follow.ChildKey(), follow.ParentKey(), data))
	}

	return NewFilter(0, buildFilterRow(follow.ParentKey(), follow.ChildKey(), data))
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
