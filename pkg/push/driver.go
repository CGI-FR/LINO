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
	"fmt"

	"github.com/rs/zerolog/log"
)

// Push write rows to target table
func Push(ri RowIterator, destination DataDestination, plan Plan, mode Mode, commitSize uint, disableConstraints bool, catchError RowWriter) *Error {
	err1 := destination.Open(plan, mode, disableConstraints)
	if err1 != nil {
		return err1
	}
	defer destination.Close()
	defer ri.Close()

	Reset()

	i := uint(0)
	for ri.Next() {
		row := ri.Value()

		err2 := pushRow(*row, destination, plan.FirstTable(), plan, mode)
		if err2 != nil {
			err4 := catchError.Write(*row)
			if err4 != nil {
				return &Error{Description: fmt.Sprintf("%s (%s)", err2.Error(), err4.Error())}
			}
			log.Warn().Msg(fmt.Sprintf("Error catched : %s", err2.Error()))
		}
		i++
		if i%commitSize == 0 {
			log.Info().Msg("Intermediate commit")
			errCommit := destination.Commit()
			if errCommit != nil {
				return errCommit
			}
			IncCommitsCount()
		}
		IncInputLinesCount()
	}
	pushStats := Compute()
	log.Info().RawJSON("stats", pushStats.ToJSON()).Msg("Push over")

	if ri.Error() != nil {
		return ri.Error()
	}

	log.Info().Msg("End of stream")
	return nil
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
				log.Debug().Msg(fmt.Sprintf("null relation for key %s", name))

			default:
				log.Error().Msg(fmt.Sprintf("key = %s", name))
				log.Error().Msg(fmt.Sprintf("type = %T", val))
				log.Error().Msg(fmt.Sprintf("val = %s", val))

				return frow, frel, fInverseRel, &Error{Description: fmt.Sprintf("%v is not a array", val)}
			}
		} else {
			frow[name] = val
		}
	}
	return frow, frel, fInverseRel, nil
}

// pushRow push a row in a specific table
func pushRow(row Row, ds DataDestination, table Table, plan Plan, mode Mode) *Error {
	frow, frel, fInverseRel, err1 := FilterRelation(row, plan.RelationsFromTable(table))

	if err1 != nil {
		return err1
	}

	rw, err2 := ds.RowWriter(table)
	if err2 != nil {
		return err2
	}

	if mode == Delete {
		// remove children first
		for relName, subArray := range fInverseRel {
			for _, subRow := range subArray {
				rel := plan.RelationsFromTable(table)[relName]
				err5 := pushRow(subRow, ds, rel.OppositeOf(table), plan, mode)
				if err5 != nil {
					return err5
				}
			}
		}

		// Current table
		err3 := rw.Write(frow)

		IncDeletedLinesCount(table.Name())

		if err3 != nil {
			return err3
		}

		// and parents
		for relName, subRow := range frel {
			rel := plan.RelationsFromTable(table)[relName]
			err4 := pushRow(subRow, ds, rel.OppositeOf(table), plan, mode)
			if err4 != nil {
				return err4
			}
		}
	} else {
		// insert parent first
		for relName, subRow := range frel {
			rel := plan.RelationsFromTable(table)[relName]
			err4 := pushRow(subRow, ds, rel.OppositeOf(table), plan, mode)
			if err4 != nil {
				return err4
			}
		}

		// current
		err3 := rw.Write(frow)

		IncCreatedLinesCount(table.Name())

		if err3 != nil {
			return err3
		}

		// and children
		for relName, subArray := range fInverseRel {
			for _, subRow := range subArray {
				rel := plan.RelationsFromTable(table)[relName]
				err5 := pushRow(subRow, ds, rel.OppositeOf(table), plan, mode)
				if err5 != nil {
					return err5
				}
			}
		}
	}

	return nil
}
