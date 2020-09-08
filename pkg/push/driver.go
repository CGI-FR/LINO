package push

import "fmt"

var logger Logger = Nologger{}

// Push write rows to target table
func Push(ri RowIterator, destination DataDestination, plan Plan, mode Mode, commitSize uint, disableConstraints bool) *Error {
	err1 := destination.Open(plan, mode, disableConstraints)
	if err1 != nil {
		return err1
	}
	defer destination.Close()

	i := uint(0)
	for {
		row, stop := ri.NextRow()
		if stop != nil {
			logger.Info("End of stream")
			return nil
		}
		err2 := pushRow(*row, destination, plan.FirstTable(), plan)
		if err2 != nil {
			return err2
		}
		i++
		if i%commitSize == 0 {
			logger.Info("Intermediate commit")
			err3 := destination.Commit()
			if err2 != nil {
				return err3
			}
		}
	}
}

// FilterRelation split values and relations to follow
func FilterRelation(row Row, relations map[string]Relation) (Row, map[string]Row, map[string][]Row, *Error) {
	frow := Row{}
	frel := map[string]Row{}
	fInverseRel := map[string][]Row{}

	for name, val := range row {
		if rel, ok := relations[name]; ok {
			switch tv := val.(type) {
			case map[string]interface{}:
				sr := Row{}
				for k, v := range tv {
					sr[k] = v
				}

				frel[rel.Name()] = sr
			case []interface{}:
				sa := []Row{}
				for _, srValue := range tv {
					var srMap map[string]interface{}
					if srMap, ok = srValue.(map[string]interface{}); !ok {
						return frow, frel, fInverseRel, &Error{Description: fmt.Sprintf("%v is not a map", val)}
					}
					sr := Row{}
					for k, v := range srMap {
						sr[k] = v
					}
					sa = append(sa, sr)
				}
				fInverseRel[rel.Name()] = sa

			case nil:
				logger.Debug(fmt.Sprintf("null relation for key %s", name))

			default:
				logger.Error(fmt.Sprintf("key = %s", name))
				logger.Error(fmt.Sprintf("type = %T", val))
				logger.Error(fmt.Sprintf("val = %s", val))

				return frow, frel, fInverseRel, &Error{Description: fmt.Sprintf("%v is not a array", val)}
			}
		} else {
			frow[name] = val
		}
	}
	return frow, frel, fInverseRel, nil
}

// pushRow push a row in a specific table
func pushRow(row Row, ds DataDestination, table Table, plan Plan) *Error {
	frow, frel, fInverseRel, err1 := FilterRelation(row, plan.RelationsFromTable(table))

	if err1 != nil {
		return err1
	}

	rw, err2 := ds.RowWriter(table)
	if err2 != nil {
		return err2
	}

	err3 := rw.Write(frow)

	if err3 != nil {
		return err3
	}

	for relName, subRow := range frel {
		rel := plan.RelationsFromTable(table)[relName]
		err4 := pushRow(subRow, ds, rel.OppositeOf(table), plan)
		if err4 != nil {
			return err4
		}
	}

	for relName, subArray := range fInverseRel {
		for _, subRow := range subArray {
			rel := plan.RelationsFromTable(table)[relName]
			err5 := pushRow(subRow, ds, rel.OppositeOf(table), plan)
			if err5 != nil {
				return err5
			}
		}
	}

	return nil
}

// SetLogger if needed, default no logger
func SetLogger(l Logger) {
	logger = l
}
