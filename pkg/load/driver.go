package load

import "fmt"

var logger Logger = Nologger{}

// Load write rows to target table
func Load(ri RowIterator, destination DataDestination, plan Plan, mode Mode) *Error {
	err1 := destination.Open(plan, mode)
	if err1 != nil {
		return err1
	}
	defer destination.Close()

	for {
		row, stop := ri.NextRow()
		if stop != nil {
			logger.Info("End of stream")
			return nil
		}
		err2 := loadRow(*row, destination, plan.FirstTable(), plan)
		if err2 != nil {
			return err2
		}
	}
}

// filterRelation split values and relations to follow
func filterRelation(row Row, relations map[string]Relation) (Row, map[Relation]Row, map[Relation][]Row, *Error) {
	frow := Row{}
	frel := map[Relation]Row{}
	fInverseRel := map[Relation][]Row{}

	for name, val := range row {
		if rel, ok := relations[name]; ok {
			switch tv := val.(type) {
			case map[string]interface{}:
				sr := Row{}
				for k, v := range tv {
					if vv, ok := v.(Value); !ok {
						logger.Trace(fmt.Sprintf("k = %s", k))
						logger.Trace(fmt.Sprintf("t = %T", v))
						logger.Trace(fmt.Sprintf("v = %s", v))
					} else {
						sr[k] = vv
					}
				}
				frel[rel] = sr
			case []interface{}:
				sa := []Row{}
				for _, srValue := range tv {
					sr := Row{}
					if srMap, ok := srValue.(map[string]interface{}); !ok {
						return frow, frel, fInverseRel, &Error{Description: fmt.Sprintf("%v is not a map", val)}
					} else {
						for k, v := range srMap {
							if vv, ok := v.(Value); !ok {
								logger.Trace(fmt.Sprintf("k = %s", k))
								logger.Trace(fmt.Sprintf("t = %T", v))
								logger.Trace(fmt.Sprintf("v = %s", v))
							} else {
								sr[k] = vv
							}

							sa = append(sa, sr)
						}
					}
				}
				fInverseRel[rel] = sa
			case []map[string]interface{}:
				sa := []Row{}
				for _, srValue := range tv {
					sr := Row{}

					for k, v := range srValue {
						if vv, ok := v.(Value); !ok {
							logger.Trace(fmt.Sprintf("k = %s", k))
							logger.Trace(fmt.Sprintf("t = %T", v))
							logger.Trace(fmt.Sprintf("v = %s", v))
						} else {
							sr[k] = vv
						}

						sa = append(sa, sr)
					}
				}
				fInverseRel[rel] = sa
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

// loadRow load a row in a specific table
func loadRow(row Row, ds DataDestination, table Table, plan Plan) *Error {
	frow, frel, fInverseRel, err1 := filterRelation(row, plan.RelationsFromTable(table))

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

	for rel, subRow := range frel {
		err4 := loadRow(subRow, ds, rel.OppositeOf(table), plan)
		if err4 != nil {
			return err4
		}
	}

	for rel, subArray := range fInverseRel {
		for _, subRow := range subArray {
			err5 := loadRow(subRow, ds, rel.OppositeOf(table), plan)
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
