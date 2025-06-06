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

	over "github.com/adrienaury/zeromdc"
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
	idStorageFactory     func(string, string) id.Storage
	dataSourceFactories  map[string]pull.DataSourceFactory
	pullExporterFactory  func(io.Writer) pull.RowExporter
	rowReaderFactory     func(io.ReadCloser) pull.RowReader
	keyStoreFactory      func(io.ReadCloser, []string) (pull.KeyStore, error)
)

var traceListener pull.TraceListener

// Inject dependencies
func Inject(
	dbas dataconnector.Storage,
	rs relation.Storage,
	ts table.Storage,
	idsf func(string, string) id.Storage,
	dsfmap map[string]pull.DataSourceFactory,
	exporterFactory func(io.Writer) pull.RowExporter,
	rrf func(io.ReadCloser) pull.RowReader,
	ksf func(io.ReadCloser, []string) (pull.KeyStore, error),
	tl pull.TraceListener,
) {
	dataconnectorStorage = dbas
	relStorage = rs
	tabStorage = ts
	idStorageFactory = idsf
	dataSourceFactories = dsfmap
	pullExporterFactory = exporterFactory
	rowReaderFactory = rrf
	keyStoreFactory = ksf
	traceListener = tl
}

// NewCommand implements the cli pull command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	// local flags
	var distinct bool
	var limit uint
	var filefilter string
	var fileexclude string
	var table string
	var ingressDescriptor string
	var where string
	var initialFilters map[string]string
	var diagnostic bool
	var filters pull.RowReader
	var parallel uint

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
				Bool("distinct", distinct).
				Str("filter-from-file", filefilter).
				Str("exclude-from-file", fileexclude).
				Str("table", table).
				Str("where", where).
				Uint("parallel", parallel).
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

			plan, start, startSelect, e2 := getPullerPlan(idStorageFactory(table, ingressDescriptor))
			if e2 != nil {
				fmt.Fprintln(err, e2.Error())
				os.Exit(1)
			}

			log.Debug().Interface("start", start).Msg("pull plan is complete")

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
				log.Trace().Str("file", filefilter).Msg("reading file")
			}

			var filtersEx pull.KeyStore
			if len(fileexclude) > 0 {
				filterReader, e3 := os.Open(fileexclude)
				if e3 != nil {
					fmt.Fprintln(err, e3.Error())
					os.Exit(1)
				}
				filtersEx, e3 = keyStoreFactory(filterReader, start.Keys)
				if e3 != nil {
					fmt.Fprintln(err, e3.Error())
					os.Exit(1)
				}
				log.Trace().Str("file", fileexclude).Msg("reading file")
			}

			row := pull.Row{}
			for column, value := range initialFilters {
				row[column] = value
			}
			filter := pull.Filter{
				Limit:    limit,
				Values:   row,
				Where:    where,
				Distinct: distinct,
			}

			puller := pull.NewPullerParallel(plan, datasource, pullExporterFactory(out), tracer, parallel)
			if e3 := puller.Pull(start, filter, startSelect, filters, filtersEx); e3 != nil {
				log.Fatal().AnErr("error", e3).Msg("Fatal error stop the pull command")
				os.Exit(1)
			}

			duration := time.Since(startTime)
			over.MDC().Set("duration", duration)
			stats := pull.Compute()
			pull.SetDuration(duration)
			over.MDC().Set("stats", stats.ToJSON())
		},
	}
	cmd.Flags().UintVarP(&limit, "limit", "l", 1, "limit the number of results")
	cmd.Flags().StringToStringVarP(&initialFilters, "filter", "f", map[string]string{}, "filter of start table")
	cmd.Flags().BoolVarP(&diagnostic, "diagnostic", "d", false, "Set diagnostic debug on")
	cmd.Flags().BoolVarP(&distinct, "distinct", "D", false, "select distinct values from start table")
	cmd.Flags().StringVarP(&filefilter, "filter-from-file", "F", "", "Use file to filter start table")
	cmd.Flags().StringVarP(&fileexclude, "exclude-from-file", "X", "", "Use file to filter out start table")
	cmd.Flags().StringVarP(&table, "table", "t", "", "pull content of table without relations instead of ingress descriptor definition")
	cmd.Flags().StringVarP(&where, "where", "w", "", "Advanced SQL where clause to filter")
	cmd.Flags().StringVarP(&ingressDescriptor, "ingress-descriptor", "i", "ingress-descriptor.yaml", "pull content using ingress descriptor definition")
	cmd.Flags().UintVarP(&parallel, "parallel", "p", 1, "number of parallel workers")
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}

func getDataSource(dataconnectorName string, out io.Writer) (pull.DataSource, error) {
	alias, e1 := dataconnector.Get(dataconnectorStorage, dataconnectorName)
	if e1 != nil {
		return nil, e1
	}
	if alias == nil {
		return nil, fmt.Errorf("Data Connector %s not found", dataconnectorName)
	}

	u := urlbuilder.BuildURL(alias, out)

	datasourceFactory, ok := dataSourceFactories[u.UnaliasedDriver]
	if !ok {
		return nil, fmt.Errorf("no datasource found for database type")
	}

	return datasourceFactory.New(u.URL.String(), alias.Schema), nil
}

func getPullerPlan(idStorage id.Storage) (pull.Plan, pull.Table, []string, error) {
	pp, err1 := id.GetPullerPlan(idStorage)
	if err1 != nil {
		return pull.Plan{}, pull.Table{}, []string{}, err1
	}

	relations, err2 := relStorage.List()
	if err2 != nil {
		return pull.Plan{}, pull.Table{}, []string{}, err2
	}

	tables, err3 := tabStorage.List()
	if err3 != nil {
		return pull.Plan{}, pull.Table{}, []string{}, err3
	}

	builder := newBuilder(pp, relations, tables)
	plan, startTable, err4 := builder.plan()
	if err4 != nil {
		return pull.Plan{}, pull.Table{}, []string{}, err4
	}

	// Check startTable existe in table.yaml
	tableExiste := false
	for _, table := range tables {
		if table.Name == string(startTable.Name) {
			tableExiste = true

			break
		}
	}

	if !tableExiste {
		err5 := fmt.Errorf("Table '%s' does not exist in table.yaml", string(startTable.Name))
		return pull.Plan{}, pull.Table{}, []string{}, err5
	}

	return plan, startTable, pp.Select(), nil
}
