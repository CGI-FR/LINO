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
	"io"
	"os"
	"time"

	over "github.com/Trendyol/overlog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/cgi-fr/lino/internal/app/urlbuilder"
	"github.com/cgi-fr/lino/pkg/dataconnector"
	"github.com/cgi-fr/lino/pkg/id"
	"github.com/cgi-fr/lino/pkg/pull"
	"github.com/cgi-fr/lino/pkg/relation"
	"github.com/cgi-fr/lino/pkg/table"
)

var (
	dataconnectorStorage dataconnector.Storage
	relStorage           relation.Storage
	tabStorage           table.Storage
	idStorageFactory     func(string) id.Storage
	dataSourceFactories  map[string]pull.DataSourceFactory
	pullExporterFactory  func(io.Writer) pull.RowExporter
	rowReaderFactory     func(io.ReadCloser) pull.RowReader
)

var traceListener pull.TraceListener

// Inject dependencies
func Inject(
	dbas dataconnector.Storage,
	rs relation.Storage,
	ts table.Storage,
	idsf func(string) id.Storage,
	dsfmap map[string]pull.DataSourceFactory,
	exporterFactory func(io.Writer) pull.RowExporter,
	rrf func(io.ReadCloser) pull.RowReader,
	tl pull.TraceListener) {
	dataconnectorStorage = dbas
	relStorage = rs
	tabStorage = ts
	idStorageFactory = idsf
	dataSourceFactories = dsfmap
	pullExporterFactory = exporterFactory
	rowReaderFactory = rrf
	traceListener = tl
}

// NewCommand implements the cli pull command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	// local flags
	var limit uint
	var filefilter string
	var table string
	var where string
	var initialFilters map[string]string
	var diagnostic bool
	var filters pull.RowReader

	cmd := &cobra.Command{
		Use:     "pull [DB Alias Name]",
		Short:   "Pull data from a database",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s pull mydatabase --limit 1", fullName),
		Args:    cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			log.Info().
				Uint("limit", limit).
				Interface("filter", initialFilters).
				Bool("diagnostic", diagnostic).
				Str("filter-from-file", filefilter).
				Str("table", table).
				Str("where", where).
				Msg("Pull mode")
		},
		Run: func(cmd *cobra.Command, args []string) {
			over.MDC().Set("action", "pull")
			over.SetGlobalFields([]string{"action"})

			startTime := time.Now()

			datasource, e1 := getDataSource(args[0], out)
			if e1 != nil {
				fmt.Fprintln(err, e1.Error())
				os.Exit(1)
			}

			plan, e2 := getPullerPlan(initialFilters, limit, where, idStorageFactory(table))
			if e2 != nil {
				fmt.Fprintln(err, e2.Error())
				os.Exit(1)
			}

			var tracer pull.TraceListener

			tracer = pull.NoTraceListener{}

			if diagnostic {
				tracer = traceListener
			}

			switch filefilter {
			case "":
				filters = pull.NewOneEmptyRowReader()
			case "-":
				filters = rowReaderFactory(in)
			default:
				filterReader, e3 := os.Open(filefilter)
				if e3 != nil {
					fmt.Fprintln(err, e3.Error())
					os.Exit(1)
				}
				filters = rowReaderFactory(filterReader)
			}
			e3 := pull.Pull(plan, filters, datasource, pullExporterFactory(out), tracer)
			if e3 != nil {
				fmt.Fprintln(err, e3.Error())
				os.Exit(1)
			}

			duration := time.Since(startTime)
			over.MDC().Set("duration", duration)
			stats := pull.Compute()
			over.MDC().Set("stats", stats.ToJSON())
			over.SetGlobalFields([]string{"duration"})
		},
	}
	cmd.Flags().UintVarP(&limit, "limit", "l", 1, "limit the number of results")
	cmd.Flags().StringToStringVarP(&initialFilters, "filter", "f", map[string]string{}, "filter of start table")
	cmd.Flags().BoolVarP(&diagnostic, "diagnostic", "d", false, "Set diagnostic debug on")
	cmd.Flags().StringVarP(&filefilter, "filter-from-file", "F", "", "Use file to filter start table")
	cmd.Flags().StringVarP(&table, "table", "t", "", "pull content of table without relations instead of ingress descriptor definition")
	cmd.Flags().StringVarP(&where, "where", "w", "", "Advanced SQL where clause to filter")
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}

func getDataSource(dataconnectorName string, out io.Writer) (pull.DataSource, *pull.Error) {
	alias, e1 := dataconnector.Get(dataconnectorStorage, dataconnectorName)
	if e1 != nil {
		return nil, &pull.Error{Description: e1.Error()}
	}
	if alias == nil {
		return nil, &pull.Error{Description: fmt.Sprintf("Data Connector %s not found", dataconnectorName)}
	}

	u := urlbuilder.BuildURL(alias, out)

	datasourceFactory, ok := dataSourceFactories[u.Unaliased]
	if !ok {
		return nil, &pull.Error{Description: "no datasource found for database type"}
	}

	return datasourceFactory.New(u.URL.String(), alias.Schema), nil
}

func getPullerPlan(initialFilters map[string]string, limit uint, where string, idStorage id.Storage) (pull.Plan, *pull.Error) {
	ep, err1 := id.GetPullerPlan(idStorage)
	if err1 != nil {
		return nil, &pull.Error{Description: err1.Error()}
	}

	relations, err2 := relStorage.List()
	if err2 != nil {
		return nil, &pull.Error{Description: err2.Error()}
	}

	tables, err3 := tabStorage.List()
	if err3 != nil {
		return nil, &pull.Error{Description: err3.Error()}
	}
	var filter pull.Filter

	stepList, err4 := getStepList(ep, relations, tables)

	if err4 != nil {
		return nil, &pull.Error{Description: err4.Error()}
	}

	row := pull.Row{}
	for column, value := range initialFilters {
		row[column] = value
	}
	filter = pull.NewFilter(limit, row, where)

	return pull.NewPlan(filter, stepList), nil
}

func getStepList(ep id.PullerPlan, relations []relation.Relation, tables []table.Table) (pull.StepList, error) {
	rmap := map[string]relation.Relation{}
	for _, relation := range relations {
		rmap[relation.Name] = relation
	}

	tmap := map[string]table.Table{}
	for _, table := range tables {
		tmap[table.Name] = table
	}

	smap := []id.Step{}
	for idx := uint(0); idx < ep.Len(); idx++ {
		smap = append(smap, ep.Step(idx))
	}

	log.Debug().Msg(fmt.Sprintf("there is %v step(s) to build", ep.Len()))

	converter := epToStepListConverter{
		rmap:   rmap,
		tmap:   tmap,
		smap:   smap,
		exrmap: map[string]pull.Relation{},
		extmap: map[string]pull.Table{},
		exsmap: map[uint]pull.Step{},
	}
	steps, err := converter.getSteps()
	if err != nil {
		return nil, err
	}
	log.Debug().Msg(fmt.Sprintf("finished building %v step(s) with success", ep.Len()))
	return steps, nil
}

type epToStepListConverter struct {
	rmap map[string]relation.Relation
	tmap map[string]table.Table
	smap []id.Step

	exrmap map[string]pull.Relation
	extmap map[string]pull.Table
	exsmap map[uint]pull.Step
}

func (c epToStepListConverter) getTable(name string) pull.Table {
	if extable, ok := c.extmap[name]; ok {
		return extable
	}

	table, ok := c.tmap[name]
	if !ok {
		log.Warn().Msg(fmt.Sprintf("missing table %v in tables.yaml", name))
		return pull.NewTable(name, []string{})
	}

	log.Trace().Msg(fmt.Sprintf("building table %v", table))

	return pull.NewTable(table.Name, table.Keys)
}

func (c epToStepListConverter) getRelation(name string) (pull.Relation, error) {
	if exrelation, ok := c.exrmap[name]; ok {
		return exrelation, nil
	}

	if name == "" {
		return pull.NewRelation(name, nil, nil, []string{}, []string{}), nil
	}

	relation, ok := c.rmap[name]
	if !ok {
		err := fmt.Errorf("missing relation '%s' in relations.yaml", name)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	log.Trace().Msg(fmt.Sprintf("building relation %v", relation))

	return pull.NewRelation(
		relation.Name,
		c.getTable(relation.Parent.Name),
		c.getTable(relation.Child.Name),
		relation.Parent.Keys,
		relation.Child.Keys,
	), nil
}

func (c epToStepListConverter) getRelationList(relations id.IngressRelationList) (pull.RelationList, error) {
	exrelations := []pull.Relation{}
	for idx := uint(0); idx < relations.Len(); idx++ {
		rel, err := c.getRelation(relations.Relation(idx).Name())
		if err != nil {
			return nil, err
		}
		exrelations = append(exrelations, rel)
	}
	return pull.NewRelationList(exrelations), nil
}

func (c epToStepListConverter) getCycleList(cycles id.CycleList) (pull.CycleList, error) {
	excycles := []pull.Cycle{}
	for idx := uint(0); idx < cycles.Len(); idx++ {
		rel, err := c.getRelationList(cycles.Cycle(idx))
		if err != nil {
			return nil, err
		}
		excycles = append(excycles, rel)
	}
	return pull.NewCycleList(excycles), nil
}

func (c epToStepListConverter) getStepList(previousStep uint) (pull.StepList, error) {
	exsteps := []pull.Step{}
	for _, step := range c.smap {
		if step.PreviousStep() == previousStep {
			step, err := c.getStep(step.Index())
			if err != nil {
				return nil, err
			}
			exsteps = append(exsteps, step)
		}
	}
	return pull.NewStepList(exsteps), nil
}

func (c epToStepListConverter) getStep(idx uint) (pull.Step, error) {
	if exstep, ok := c.exsmap[idx]; ok {
		return exstep, nil
	}

	step := c.smap[idx-1]

	log.Trace().Msg(fmt.Sprintf("building %v", step))

	var exstep pull.Step
	rel, err := c.getRelation(step.Following().Name())
	if err != nil {
		return nil, err
	}
	relList, err := c.getRelationList(step.Relations())
	if err != nil {
		return nil, err
	}
	cycleList, err := c.getCycleList(step.Cycles())
	if err != nil {
		return nil, err
	}

	stepList, err := c.getStepList(step.Index())
	if err != nil {
		return nil, err
	}
	if step.Index() > 1 {
		exstep = pull.NewStep(
			step.Index(),
			c.getTable(step.Entry().Name()),
			rel,
			relList,
			cycleList,
			stepList,
		)
	} else {
		exstep = pull.NewStep(
			step.Index(),
			c.getTable(step.Entry().Name()),
			nil,
			relList,
			cycleList,
			stepList,
		)
	}

	c.exsmap[idx] = exstep

	return exstep, nil
}

func (c epToStepListConverter) getSteps() (pull.StepList, error) {
	exsteps := []pull.Step{}
	for _, step := range c.smap {
		step, err := c.getStep(step.Index())
		if err != nil {
			return nil, err
		}
		exsteps = append(exsteps, step)
	}
	return pull.NewStepList(exsteps), nil
}
