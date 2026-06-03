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
	"io"
	"os"
	"strings"
	"time"

	over "github.com/adrienaury/zeromdc"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/cgi-fr/lino/internal/app/urlbuilder"
	"github.com/cgi-fr/lino/pkg/dataconnector"
	"github.com/cgi-fr/lino/pkg/id"
	"github.com/cgi-fr/lino/pkg/push"
	"github.com/cgi-fr/lino/pkg/relation"
	"github.com/cgi-fr/lino/pkg/table"
)

var (
	dataconnectorStorage     dataconnector.Storage
	relStorage               relation.Storage
	tabStorage               table.Storage
	idStorageFactory         func(string, string) id.Storage
	datadestinationFactories map[string]push.DataDestinationFactory
	rowIteratorFactory       func(io.ReadCloser) push.RowIterator
	rowExporterFactory       func(io.Writer) push.RowWriter
	translator               push.Translator
	observer                 push.Observer
)

// Inject dependencies
func Inject(
	dbas dataconnector.Storage,
	rs relation.Storage,
	ts table.Storage,
	idsf func(string, string) id.Storage,
	dsfmap map[string]push.DataDestinationFactory,
	rif func(io.ReadCloser) push.RowIterator,
	ref func(io.Writer) push.RowWriter,
	trnsltor push.Translator,
	obs push.Observer,
) {
	dataconnectorStorage = dbas
	relStorage = rs
	tabStorage = ts
	idStorageFactory = idsf
	datadestinationFactories = dsfmap
	rowIteratorFactory = rif
	rowExporterFactory = ref
	translator = trnsltor
	observer = obs
}

// parseArgument get dataconnector and mode from args
func parseArguments(args []string) (string, push.Mode) {
	dcDestination := args[0]
	mode, _ := push.ParseMode("insert")

	if len(args) == 2 {
		dcDestination = args[1]
		mode, _ = push.ParseMode(args[0])
	}

	return dcDestination, mode
}

// NewCommand implements the cli pull command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	var (
		commitSize         uint
		disableConstraints bool
		catchErrors        string
		table              string
		ingressDescriptor  string
		rowExporter        push.RowWriter
		pkTranslations     map[string]string
		whereField         string
		savepoint          string
		autoTruncate       bool
		watch              bool
		logSQLTo           string
		commitTimeout      time.Duration
	)

	cmd := &cobra.Command{
		Use:     "push {<truncate>|<insert>|<update>|<delete>|<upsert>} [Data Connector Name]",
		Short:   "Push data to a database with a pushing mode (insert by default)",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s push truncate dstdatabase\n  %[1]s push dstdatabase", fullName),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return nil
			}
			if len(args) == 2 {
				if _, err := push.ParseMode(args[0]); err != nil {
					return err
				}
				return nil
			}
			return fmt.Errorf("accepts 1 or 2 args, received %d", len(args))
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			dcDestination, mode := parseArguments(args)

			log.Info().
				Str("dataconnector", dcDestination).
				Str("mode", mode.String()).
				Uint("commitSize", commitSize).
				Bool("disable-constraints", disableConstraints).
				Str("catch-errors", catchErrors).
				Str("table", table).
				Msg("Push mode")
		},
		Run: func(cmd *cobra.Command, args []string) {
			over.MDC().Set("action", "push")
			over.SetGlobalFields([]string{"action"})

			startTime := time.Now()

			dcDestination, mode := parseArguments(args)

			datadestination, e1 := getDataDestination(dcDestination)
			if e1 != nil {
				fmt.Fprintln(err, e1.Error())
				os.Exit(1)
			}

			if logSQLTo != "" {
				if e := datadestination.OpenSQLLogger(logSQLTo); e != nil {
					log.Warn().Err(e).Msg("error while opening SQL logger")
				}
			}

			plan, e2 := getPlan(idStorageFactory(table, ingressDescriptor), autoTruncate)
			if e2 != nil {
				fmt.Fprintln(err, e2.Error())
				os.Exit(2)
			}
			log.Debug().Msg(fmt.Sprintf("call Push with mode %s", mode))

			if catchErrors != "" {
				errorFile, e4 := os.Create(catchErrors)
				if e4 != nil {
					fmt.Fprintln(err, e4.Error())
					os.Exit(4)
				}
				defer errorFile.Close()
				rowExporter = rowExporterFactory(errorFile)
			} else {
				rowExporter = push.NoErrorCaptureRowWriter{}
			}

			if err := loadTranslator(pkTranslations); err != nil {
				log.Fatal().AnErr("error", err).Msg("Fatal error stop the push command")
				os.Exit(1)
			}

			observers := []push.Observer{}
			if watch {
				observers = append(observers, observer)
			}

			e3 := push.Push(rowIteratorFactory(in), datadestination, plan, mode, commitSize, commitTimeout, disableConstraints, rowExporter, translator, whereField, savepoint, autoTruncate, observers...)
			if e3 != nil {
				log.Fatal().AnErr("error", e3).Msg("Fatal error stop the push command")
				os.Exit(1)
			}

			duration := time.Since(startTime)
			over.MDC().Set("duration", duration)
			stats := push.Compute()
			push.SetDuration(duration)
			over.MDC().Set("stats", stats.ToJSON())
		},
	}
	cmd.Flags().UintVarP(&commitSize, "commitSize", "c", 500, "Commit size")
	cmd.Flags().DurationVar(&commitTimeout, "commit-timeout", 0, "Commit timeout (e.g. 5s, 1m). If set, a commit is triggered if no new row is received within this duration.")
	cmd.Flags().BoolVarP(&disableConstraints, "disable-constraints", "d", false, "Disable constraint during push")
	cmd.Flags().StringVarP(&catchErrors, "catch-errors", "e", "", "Catch errors and write line in file")
	cmd.Flags().StringVarP(&table, "table", "t", "", "Table to writes json")
	cmd.Flags().StringVarP(&ingressDescriptor, "ingress-descriptor", "i", "ingress-descriptor.yaml", "Ingress descriptor filename")
	cmd.Flags().StringToStringVar(&pkTranslations, "pk-translation", map[string]string{}, "list of dictionaries old value / new value for primary key update")
	cmd.Flags().StringVar(&whereField, "using-pk-field", "__usingpk__", "Name of the data field that can be used as pk for update queries")
	cmd.Flags().StringVar(&savepoint, "savepoint", "", "Name of a file to write primary keys of effectively processed lines (commit to database)")
	cmd.Flags().BoolVarP(&autoTruncate, "autotruncate", "a", false, "Automatically truncate values to the maximum length defined in table.yaml")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch statistics about pushed lines")
	cmd.Flags().StringVarP(&logSQLTo, "log-sql", "l", "", "Log SQL requests and data to specified folder (1 file per table)")
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}

func loadTranslator(pkTranslations map[string]string) error {
	fileToKeys := map[string][]push.Key{}
	fileToRows := map[string]push.RowIterator{}

	for key, filename := range pkTranslations {
		tableAndColumn := strings.SplitN(key, ".", 2)
		key := push.Key{TableName: tableAndColumn[0], ColumnName: tableAndColumn[1]}

		log.Debug().Str("table", tableAndColumn[0]).Str("column", tableAndColumn[1]).Str("file", filename).Msg("enabling key translation on primary key")

		if keys, exists := fileToKeys[filename]; exists {
			fileToKeys[filename] = append(keys, key)

			continue
		}

		fileToKeys[filename] = []push.Key{key}

		file, err := os.Open(filename)
		if err != nil {
			return err
		}

		defer file.Close()

		rowIterator := rowIteratorFactory(file)
		defer rowIterator.Close()

		fileToRows[filename] = rowIterator
	}

	for filename, rowIterator := range fileToRows {
		log.Debug().Str("file", filename).Msg("loading key translation cache from file")
		if err := translator.Load(fileToKeys[filename], rowIterator); err != nil {
			return err
		}
	}

	return nil
}

func getDataDestination(dataconnectorName string) (push.DataDestination, *push.Error) {
	alias, e1 := dataconnector.Get(dataconnectorStorage, dataconnectorName)
	if e1 != nil {
		return nil, &push.Error{Description: e1.Error()}
	}
	if alias == nil {
		return nil, &push.Error{Description: fmt.Sprintf("'%s' dataconnector not found", dataconnectorName)}
	}
	if alias.ReadOnly {
		return nil, &push.Error{Description: fmt.Sprintf("'%s' is a read only dataconnector", alias.Name)}
	}

	u := urlbuilder.BuildURL(alias, nil)

	datadestinationFactory, ok := datadestinationFactories[u.UnaliasedDriver]
	if !ok {
		return nil, &push.Error{Description: "no datadestination found for database type " + u.UnaliasedDriver}
	}

	return datadestinationFactory.New(u.URL.String(), alias.Schema), nil
}

func getPlan(idStorage id.Storage, autoTruncate bool) (push.Plan, *push.Error) {
	id, err1 := idStorage.Read()
	if err1 != nil {
		return nil, &push.Error{Description: err1.Error()}
	}

	relations, err2 := relStorage.List()
	if err2 != nil {
		return nil, &push.Error{Description: err2.Error()}
	}

	tables, err3 := tabStorage.List()
	if err3 != nil {
		return nil, &push.Error{Description: err3.Error()}
	}

	rmap := map[string]relation.Relation{}
	for _, relation := range relations {
		rmap[relation.Name] = relation
	}

	tmap := map[string]table.Table{}
	for _, table := range tables {
		tmap[table.Name] = table
	}

	converter := idToPushConverter{
		rmap:     rmap,
		tmap:     tmap,
		pushrmap: map[string]push.Relation{},
		pushtmap: map[string]push.Table{},
	}

	return converter.getPlan(id, autoTruncate), nil
}

type idToPushConverter struct {
	rmap map[string]relation.Relation
	tmap map[string]table.Table

	pushrmap map[string]push.Relation
	pushtmap map[string]push.Table
}

func (c idToPushConverter) getTable(name string, autoTruncate bool) push.Table {
	if pushtable, ok := c.pushtmap[name]; ok {
		return pushtable
	}

	table, ok := c.tmap[name]
	if !ok {
		log.Warn().Msg(fmt.Sprintf("missing table %v in tables.yaml", name))
		return push.NewTable(name, []string{}, nil)
	}

	log.Trace().Msg(fmt.Sprintf("building table %v", table))

	columns := []push.Column{}
	for _, col := range table.Columns {
		columns = append(columns, push.NewColumn(col.Name, col.Export, col.Import, col.DBInfo.Length, col.DBInfo.ByteBased, autoTruncate, col.Preserve))
	}

	return push.NewTable(table.Name, table.Keys, push.NewColumnList(columns))
}

func (c idToPushConverter) getRelation(name string, autoTruncate bool) push.Relation {
	if pushrelation, ok := c.pushrmap[name]; ok {
		return pushrelation
	}

	relation, ok := c.rmap[name]
	if !ok {
		log.Error().Err(fmt.Errorf("missing relation %v in relations.yaml", name)).Msg("")
		return push.NewRelation(name, nil, nil)
	}

	log.Trace().Msg(fmt.Sprintf("building relation %v", relation))

	return push.NewRelation(
		relation.Name,
		c.getTable(relation.Parent.Name, autoTruncate),
		c.getTable(relation.Child.Name, autoTruncate),
	)
}

func (c idToPushConverter) getPlan(idesc id.IngressDescriptor, autoTruncate bool) push.Plan {
	relations := []push.Relation{}

	activeTables, err := id.GetActiveTables(idesc)
	if err != nil {
		activeTables = id.NewTableList([]id.Table{idesc.StartTable()})
	}

	for idx := uint(0); idx < idesc.Relations().Len(); idx++ {
		rel := idesc.Relations().Relation(idx)
		if (activeTables.Contains(rel.Child().Name()) && rel.LookUpChild()) ||
			(activeTables.Contains(rel.Parent().Name()) && rel.LookUpParent()) {
			relations = append(relations, c.getRelation(rel.Name(), autoTruncate))
		}
	}

	return push.NewPlan(c.getTable(idesc.StartTable().Name(), autoTruncate), relations)
}
