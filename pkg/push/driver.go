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

package push

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

// Push write rows to target table
func Push(ri RowIterator, destination DataDestination, plan Plan, mode Mode, commitSize uint, disableConstraints bool, catchError RowWriter, translator Translator, whereField string, savepointPath string, autotruncate bool, observers ...Observer) (err *Error) {
	defer func() {
		for _, observer := range observers {
			if observer != nil {
				observer.Close()
			}
		}
	}()

	err1 := destination.Open(plan, mode, disableConstraints)
	if err1 != nil {
		return err1
	}

	defer func() {
		er1 := destination.Close()
		er2 := ri.Close()

		switch {
		case er1 != nil && er2 == nil && err == nil:
			err = er1
		case er2 != nil && er1 == nil && err == nil:
			err = er2
		case err != nil && er1 == nil && er2 == nil:
			// err = err
		case err != nil || er1 != nil || er2 != nil:
			err = &Error{Description: fmt.Sprintf("multiple errors: [%s], [%s], [%s]", err, er1, er2)}
		}
	}()

	Reset()

	committed := make([]Row, 0, commitSize)

	defer func() {
		if savepointPath != "" {
			if err := savepoint(savepointPath, committed); err != nil {
				log.Error().Msgf("Savepoint failure, %d lines committed unsaved", len(committed))
				for _, unsaved := range committed {
					log.Warn().Interface("value", unsaved).Msg("Unsaved committed value")
				}
			}
		}
	}()

	i := uint(0)
	for ri.Next() {
		row := ri.Value()

		err2 := pushRow(*row, destination, plan.FirstTable(), plan, mode, translator, whereField)
		if err2 != nil {
			err4 := catchError.Write(*row, nil)
			if err4 != nil {
				return &Error{Description: fmt.Sprintf("%s (%s)", err2.Error(), err4.Error())}
			}
			log.Warn().Msg(fmt.Sprintf("Error catched : %s", err2.Error()))
		}
		i++
		if savepointPath != "" {
			committed = append(committed, extractValues(*row, plan.FirstTable().PrimaryKey()))
		}
		if i%commitSize == 0 {
			log.Info().Msg("Intermediate commit")
			errCommit := destination.Commit()
			if errCommit != nil {
				return errCommit
			}
			if savepointPath != "" {
				if err := savepoint(savepointPath, committed); err != nil {
					return err
				}
				committed = committed[:0] // clear slice without releasing memory
			}
			IncCommitsCount()
		}
		IncInputLinesCount()
		for _, observer := range observers {
			if observer != nil {
				observer.Pushed()
			}
		}
	}

	if ri.Error() != nil {
		return ri.Error()
	}

	log.Info().Msg("End of stream")
	return nil
}

// FilterRelation split values and relations to follow
func FilterRelation(row Row, relations map[string]Relation, whereField string) (Row, Row, map[string]Row, map[string][]Row, *Error) {
	frow := Row{}
	fwhere := Row{}
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
						return frow, fwhere, frel, fInverseRel, &Error{Description: fmt.Sprintf("%v is not a map", val)}
					}
					sr := Row{}
					for k, v := range srMap {
						sr[k] = v
					}
					sa = append(sa, sr)
				}
				fInverseRel[rel.Name()] = sa

			case nil:
				log.Debug().Msg(fmt.Sprintf("null relation for key %s", name))

			default:
				log.Error().Msg(fmt.Sprintf("key = %s", name))
				log.Error().Msg(fmt.Sprintf("type = %T", val))
				log.Error().Msg(fmt.Sprintf("val = %s", val))

				return frow, fwhere, frel, fInverseRel, &Error{Description: fmt.Sprintf("%v is not a array", val)}
			}
		} else {
			if name != whereField {
				frow[name] = val
			} else if tv, ok := val.(map[string]interface{}); ok {
				for k, v := range tv {
					fwhere[k] = v
				}
			}
		}
	}
	return frow, fwhere, frel, fInverseRel, nil
}

// pushRow push a row in a specific table
func pushRow(row Row, ds DataDestination, table Table, plan Plan, mode Mode, translator Translator, whereField string) *Error {
	frow, fwhere, frel, fInverseRel, err1 := FilterRelation(row, plan.RelationsFromTable(table), whereField)
	if err1 != nil {
		return err1
	}

	// remove not imported values from frow
	if columns := table.Columns(); columns != nil {
		for i := uint(0); i < columns.Len(); i++ {
			if columns.Column(i).Import() == "no" {
				delete(frow, columns.Column(i).Name())
			}
		}
	}

	rw, err2 := ds.RowWriter(table)
	if err2 != nil {
		return err2
	}

	var where Row
	if mode == Delete || mode == Update {
		where = computeTranslatedKeys(row, table, translator)

		for key, val := range fwhere {
			where[key] = val
		}
	}

	if mode == Delete || mode == Update {
		// children first
		for relName, subArray := range fInverseRel {
			for _, subRow := range subArray {
				rel := plan.RelationsFromTable(table)[relName]
				err5 := pushRow(subRow, ds, rel.OppositeOf(table), plan, mode, translator, whereField)
				if err5 != nil {
					return err5
				}
			}
		}

		// Current table
		err3 := rw.Write(frow, where)

		IncDeletedLinesCount(table.Name())

		if err3 != nil {
			return err3
		}

		// and parents
		for relName, subRow := range frel {
			rel := plan.RelationsFromTable(table)[relName]
			err4 := pushRow(subRow, ds, rel.OppositeOf(table), plan, mode, translator, whereField)
			if err4 != nil {
				return err4
			}
		}
	} else { // Insert, Truncate
		// parent first
		for relName, subRow := range frel {
			rel := plan.RelationsFromTable(table)[relName]
			err4 := pushRow(subRow, ds, rel.OppositeOf(table), plan, mode, translator, whereField)
			if err4 != nil {
				return err4
			}
		}

		// current
		err3 := rw.Write(frow, where)

		IncCreatedLinesCount(table.Name())

		if err3 != nil {
			return err3
		}

		// and children
		for relName, subArray := range fInverseRel {
			for _, subRow := range subArray {
				rel := plan.RelationsFromTable(table)[relName]
				err5 := pushRow(subRow, ds, rel.OppositeOf(table), plan, mode, translator, whereField)
				if err5 != nil {
					return err5
				}
			}
		}
	}

	return nil
}

func computeTranslatedKeys(row Row, table Table, translator Translator) Row {
	where := Row{}

	if translator != nil {
		for _, pkname := range table.PrimaryKey() {
			newvalue := row[pkname]
			oldvalue := translator.FindValue(Key{table.Name(), pkname}, newvalue)
			where[pkname] = oldvalue
		}
	}

	return where
}

func savepoint(savepointPath string, committed []Row) *Error {
	f, err := os.OpenFile(savepointPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return &Error{Description: err.Error()}
	}
	defer f.Close()

	for _, row := range committed {
		bytes, err := json.Marshal(row)
		if err != nil {
			return &Error{Description: err.Error()}
		}

		if _, err := f.Write(append(bytes, '\n')); err != nil {
			return &Error{Description: err.Error()}
		}
	}

	return nil
}

func extractValues(row Row, keys []string) Row {
	result := Row{}
	for _, key := range keys {
		result[key] = row[key]
	}
	return result
}
