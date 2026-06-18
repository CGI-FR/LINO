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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// PushConfig holds configuration for the push operation
type PushConfig struct {
	CommitSize         uint
	CommitTimeout      time.Duration
	DisableConstraints bool
	WhereField         string
	WhereClause        string
	SavepointPath      string
	AutoTruncate       bool
}

// pushContext encapsulates the state of a push operation
type pushContext struct {
	cfg         PushConfig
	destination DataDestination
	plan        Plan
	mode        Mode
	catchError  RowWriter
	translator  Translator
	observers   []Observer

	committed  []Row
	inputCount uint
}

// Push write rows to target table
func Push(ri RowIterator, destination DataDestination, plan Plan, mode Mode, commitSize uint, commitTimeout time.Duration, disableConstraints bool, catchError RowWriter, translator Translator, whereField string, whereClause string, savepointPath string, autotruncate bool, observers ...Observer) *Error {
	cfg := PushConfig{
		CommitSize:         commitSize,
		CommitTimeout:      commitTimeout,
		DisableConstraints: disableConstraints,
		WhereField:         whereField,
		WhereClause:        whereClause,
		SavepointPath:      savepointPath,
		AutoTruncate:       autotruncate,
	}

	ctx := &pushContext{
		cfg:         cfg,
		destination: destination,
		plan:        plan,
		mode:        mode,
		catchError:  catchError,
		translator:  translator,
		observers:   observers,
		committed:   make([]Row, 0, commitSize),
	}

	return ctx.Run(ri)
}

func (ctx *pushContext) Run(ri RowIterator) (err *Error) {
	defer ctx.closeObservers()

	log.Info().
		Str("url", ctx.destination.SafeUrl()).
		Msg("Open database")

	if err := ctx.destination.Open(ctx.plan, ctx.mode, ctx.cfg.DisableConstraints, ctx.cfg.WhereClause); err != nil {
		return err
	}

	defer func() {
		err = ctx.cleanup(ri, err)
	}()

	Reset()

	// Handle savepoint on exit
	defer func() {
		if ctx.cfg.SavepointPath != "" {
			if err := savepoint(ctx.cfg.SavepointPath, ctx.committed); err != nil {
				log.Error().Msgf("Savepoint failure, %d lines committed unsaved", len(ctx.committed))
				for _, unsaved := range ctx.committed {
					log.Warn().Interface("value", unsaved).Msg("Unsaved committed value")
				}
			}
		}
	}()

	return ctx.processLoop(ri)
}

func (ctx *pushContext) closeObservers() {
	for _, observer := range ctx.observers {
		if observer != nil {
			observer.Close()
		}
	}
}

func (ctx *pushContext) cleanup(ri RowIterator, err *Error) *Error {
	er1 := ctx.destination.Close()
	er2 := ri.Close()

	// Use helper that aggregates multiple *Error into a single *Error using errors.Join
	return combineErrors(err, er1, er2)
}

func (ctx *pushContext) processLoop(ri RowIterator) *Error {
	rowChan, errChan, quit := ctx.startRowReader(ri)
	defer close(quit)

	var timer *time.Timer
	var timerCh <-chan time.Time

	if ctx.cfg.CommitTimeout > 0 {
		timer = time.NewTimer(ctx.cfg.CommitTimeout)
		// Ensure timer is stopped when we exit to avoid leaks, though we are exiting anyway.
		defer func() {
			if timer != nil {
				timer.Stop()
			}
		}()
		timerCh = timer.C
	}

loop:
	for {
		select {
		case row, ok := <-rowChan:
			if !ok {
				select {
				case e := <-errChan:
					return e
				default:
				}
				break loop
			}

			if ctx.cfg.CommitTimeout > 0 {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(ctx.cfg.CommitTimeout)
			}

			if err := ctx.processRow(row); err != nil {
				return err
			}

		case <-timerCh:
			if err := ctx.handleTimeout(); err != nil {
				return err
			}
		}
	}

	// Final commit for any remaining uncommitted rows
	if ctx.inputCount%ctx.cfg.CommitSize != 0 {
		log.Info().Msg("Final commit")
		if err := ctx.commit(); err != nil {
			return err
		}
	}

	log.Info().Msg("End of stream")
	return nil
}

func (ctx *pushContext) startRowReader(ri RowIterator) (<-chan *Row, <-chan *Error, chan struct{}) {
	rowChan := make(chan *Row)
	errChan := make(chan *Error, 1)
	quit := make(chan struct{})

	go func() {
		defer close(rowChan)
		for ri.Next() {
			val := ri.Value()
			// Shallow copy to avoid race conditions
			newRow := make(Row, len(*val))
			for k, v := range *val {
				newRow[k] = v
			}
			select {
			case rowChan <- &newRow:
			case <-quit:
				return
			}
		}
		if e := ri.Error(); e != nil {
			select {
			case errChan <- e:
			case <-quit:
				return
			}
		}
	}()

	return rowChan, errChan, quit
}

func (ctx *pushContext) processRow(row *Row) *Error {
	err := pushRow(*row, ctx.destination, ctx.plan.FirstTable(), ctx.plan, ctx.mode, ctx.translator, ctx.cfg.WhereField)
	if err != nil {
		if errWrite := ctx.catchError.Write(*row, nil); errWrite != nil {
			return &Error{Description: fmt.Sprintf("%s (%s)", err.Error(), errWrite.Error())}
		}
		log.Warn().Msg(fmt.Sprintf("Error catched : %s", err.Error()))
	}

	ctx.inputCount++
	if ctx.cfg.SavepointPath != "" {
		ctx.committed = append(ctx.committed, extractValues(*row, ctx.plan.FirstTable().PrimaryKey()))
	}

	if ctx.inputCount%ctx.cfg.CommitSize == 0 {
		log.Info().Msg("Intermediate commit")
		if err := ctx.commit(); err != nil {
			return err
		}
	}

	IncInputLinesCount()
	for _, observer := range ctx.observers {
		if observer != nil {
			observer.Pushed()
		}
	}
	return nil
}

func (ctx *pushContext) handleTimeout() *Error {
	if ctx.inputCount%ctx.cfg.CommitSize != 0 {
		log.Info().Msg("Timeout commit")
		return ctx.commit()
	}
	return nil
}

func (ctx *pushContext) commit() *Error {
	if err := ctx.destination.Commit(); err != nil {
		return err
	}
	if ctx.cfg.SavepointPath != "" {
		if err := savepoint(ctx.cfg.SavepointPath, ctx.committed); err != nil {
			// Restore previous behavior: log savepoint failures but do not make them fatal.
			log.Error().Msgf("Savepoint failure, %d lines committed unsaved: %v", len(ctx.committed), err)
			for _, unsaved := range ctx.committed {
				log.Warn().Interface("value", unsaved).Msg("Unsaved committed value")
			}
			// clear committed slice (we consider them committed to the destination even if savepoint failed)
			ctx.committed = ctx.committed[:0]
		} else {
			ctx.committed = ctx.committed[:0]
		}
	}
	IncCommitsCount()
	return nil
}

// combineErrors aggregates multiple *Error values into a single *Error.
// It uses errors.Join for multi-error aggregation and preserves single errors.
func combineErrors(errs ...*Error) *Error {
	var nonNil []error
	for _, e := range errs {
		if e != nil {
			nonNil = append(nonNil, e)
		}
	}
	switch len(nonNil) {
	case 0:
		return nil
	case 1:
		// If it's already our *Error type, return it as-is
		if single, ok := nonNil[0].(*Error); ok {
			return single
		}
		return &Error{Description: nonNil[0].Error()}
	default:
		joined := errors.Join(nonNil...)
		return &Error{Description: joined.Error()}
	}
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
	if mode == Delete || mode == Update || mode == Upsert {
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
