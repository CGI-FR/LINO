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
	"context"
	"fmt"
	"sync"

	over "github.com/adrienaury/zeromdc"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
)

type pullerParallel struct {
	puller

	nbworkers uint
	inChan    chan Row
	errChan   chan error
	outChan   chan ExportedRow
	errors    []error
	stats     chan stats
}

func NewPullerParallel(plan Plan, datasource DataSource, exporter RowExporter, diagnostic TraceListener, nbworkers uint) Puller { //nolint:lll
	puller := &puller{
		graph:      plan.buildGraph(),
		datasource: datasource,
		exporter:   exporter,
		diagnostic: diagnostic,
	}

	if nbworkers > 1 {
		return &pullerParallel{
			puller:    *puller,
			nbworkers: nbworkers,
			inChan:    nil,
			errChan:   nil,
			outChan:   nil,
			errors:    nil,
			stats:     nil,
		}
	}

	return puller
}

func (p *pullerParallel) Pull(start Table, filter Filter, filterCohort RowReader) error {
	start = p.graph.addMissingColumns(start)

	if err := p.datasource.Open(); err != nil {
		return fmt.Errorf("%w", err)
	}

	defer p.datasource.Close()

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

	p.inChan = make(chan Row)
	p.errChan = make(chan error)
	p.outChan = make(chan ExportedRow)
	p.errors = []error{}
	p.stats = make(chan stats, p.nbworkers)

	wg := &sync.WaitGroup{}

	for i := uint(0); i < p.nbworkers; i++ {
		wg.Add(1)

		go p.worker(context.Background(), wg, i, start)
	}

	done := make(chan struct{})
	go p.collect(done)
	Reset()
	for _, f := range filters {
		IncFiltersCount()
		reader, err := p.datasource.RowReader(start, f)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		for reader.Next() {
			IncLinesPerStepCount(string(start.Name))
			p.inChan <- reader.Value()
		}

		if reader.Error() != nil {
			return fmt.Errorf("%w", reader.Error())
		}
	}

	close(p.inChan)

	for idx := 0; idx < int(p.nbworkers); idx++ {
		MutualizeStats(<-p.stats)
	}

	wg.Wait()
	close(p.errChan)
	close(p.outChan)
	close(p.stats)

	<-done

	if len(p.errors) > 0 {
		return multierror.Append(p.errors[0], p.errors[1:]...)
	}

	return nil
}

func (p *pullerParallel) worker(ctx context.Context, wg *sync.WaitGroup, id uint, start Table) {
	Reset()

	over.MDC().Set("workerid", id)
	over.AddGlobalFields("workerid")
	log.Debug().Msg("start worker")
	defer wg.Done()
	defer log.Debug().Msg("end worker")

LOOP:
	for p.inChan != nil {
		log.Debug().Msg("waiting for row")
		select {
		case row, ok := <-p.inChan:
			if !ok {
				break LOOP
			}
			log.Debug().Msg("received row")

			out := start.export(row)
			err := p.pull(start, out)
			if err != nil {
				p.errChan <- err
			} else {
				p.outChan <- out
				log.Debug().Msg("exported row")
			}

		case <-ctx.Done():
			break LOOP
		}
	}

	p.stats <- *getStats()
}

func (p *pullerParallel) collect(done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

LOOP:
	for p.errChan != nil || p.outChan != nil {
		log.Debug().Msg("waiting for error or result")
		select {
		case err, ok := <-p.errChan:
			if !ok {
				p.errChan = nil

				continue LOOP
			}

			log.Error().Err(err).Msg("received error")

			p.errors = append(p.errors, err)
		case result, ok := <-p.outChan:
			if !ok {
				p.outChan = nil

				continue LOOP
			}

			log.Debug().Msg("received result")

			if err := p.exporter.Export(result); err != nil {
				p.errors = append(p.errors, err)
			}
		}
	}
}
