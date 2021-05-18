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
	idStorageFactory         func(string) id.Storage
	datadestinationFactories map[string]push.DataDestinationFactory
	rowIteratorFactory       func(io.ReadCloser) push.RowIterator
	rowExporterFactory       func(io.Writer) push.RowWriter
)

// Inject dependencies
func Inject(
	dbas dataconnector.Storage,
	rs relation.Storage,
	ts table.Storage,
	idsf func(string) id.Storage,
	dsfmap map[string]push.DataDestinationFactory,
	rif func(io.ReadCloser) push.RowIterator,
	ref func(io.Writer) push.RowWriter,
) {
	dataconnectorStorage = dbas
	relStorage = rs
	tabStorage = ts
	idStorageFactory = idsf
	datadestinationFactories = dsfmap
	rowIteratorFactory = rif
	rowExporterFactory = ref
}

// NewCommand implements the cli pull command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	var (
		commitSize         uint
		disableConstraints bool
		catchErrors        string
		table              string
		rowExporter        push.RowWriter
	)

	cmd := &cobra.Command{
		Use:     "push {<truncate>|<insert>|<update>|<delete>} [Data Connector Name]",
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
		Run: func(cmd *cobra.Command, args []string) {
			dcDestination := args[0]
			mode, _ := push.ParseMode("insert")

			if len(args) == 2 {
				dcDestination = args[1]
				mode, _ = push.ParseMode(args[0])
			}

			datadestination, e1 := getDataDestination(dcDestination)
			if e1 != nil {
				fmt.Fprintln(err, e1.Error())
				os.Exit(1)
			}

			plan, e2 := getPlan(idStorageFactory(table))
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
			e3 := push.Push(rowIteratorFactory(in), datadestination, plan, mode, commitSize, disableConstraints, rowExporter)
			if e3 != nil {
				fmt.Fprintln(err, e3.Error())
				os.Exit(1)
			}
		},
	}
	cmd.Flags().UintVarP(&commitSize, "commitSize", "c", 500, "Commit size")
	cmd.Flags().BoolVarP(&disableConstraints, "disable-constraints", "d", false, "Disable constraint during push")
	cmd.Flags().StringVarP(&catchErrors, "catch-errors", "e", "", "Catch errors and write line in file")
	cmd.Flags().StringVarP(&table, "table", "t", "", "Table to writes json")
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
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

	datadestinationFactory, ok := datadestinationFactories[u.Unaliased]
	if !ok {
		return nil, &push.Error{Description: "no datadestination found for database type " + u.Unaliased}
	}

	return datadestinationFactory.New(u.URL.String(), alias.Schema), nil
}

func getPlan(idStorage id.Storage) (push.Plan, *push.Error) {
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

	return converter.getPlan(id), nil
}

type idToPushConverter struct {
	rmap map[string]relation.Relation
	tmap map[string]table.Table

	pushrmap map[string]push.Relation
	pushtmap map[string]push.Table
}

func (c idToPushConverter) getTable(name string) push.Table {
	if pushtable, ok := c.pushtmap[name]; ok {
		return pushtable
	}

	table, ok := c.tmap[name]
	if !ok {
		log.Warn().Msg(fmt.Sprintf("missing table %v in tables.yaml", name))
		return push.NewTable(name, []string{})
	}

	log.Trace().Msg(fmt.Sprintf("building table %v", table))

	return push.NewTable(table.Name, table.Keys)
}

func (c idToPushConverter) getRelation(name string) push.Relation {
	if pushrelation, ok := c.pushrmap[name]; ok {
		return pushrelation
	}

	relation, ok := c.rmap[name]
	if !ok {
		log.Error().Msg(fmt.Sprintf("missing relation %v in relations.yaml", name))
		return push.NewRelation(name, nil, nil)
	}

	log.Trace().Msg(fmt.Sprintf("building relation %v", relation))

	return push.NewRelation(
		relation.Name,
		c.getTable(relation.Parent.Name),
		c.getTable(relation.Child.Name),
	)
}

func (c idToPushConverter) getPlan(id id.IngressDescriptor) push.Plan {
	relations := []push.Relation{}

	for idx := uint(0); idx < id.Relations().Len(); idx++ {
		rel := id.Relations().Relation(idx)
		relations = append(relations, c.getRelation(rel.Name()))
	}

	return push.NewPlan(c.getTable(id.StartTable().Name()), relations)
}
