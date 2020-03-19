package load

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xo/dburl"

	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
	"makeit.imfr.cgi.com/lino/pkg/id"
	"makeit.imfr.cgi.com/lino/pkg/load"
	"makeit.imfr.cgi.com/lino/pkg/relation"
	"makeit.imfr.cgi.com/lino/pkg/table"
)

var dataconnectorStorage dataconnector.Storage
var relStorage relation.Storage
var tabStorage table.Storage
var idStorage id.Storage
var datadestinationFactories map[string]load.DataDestinationFactory
var rowIterator load.RowIterator

var logger load.Logger = load.Nologger{}

// SetLogger if needed, default no logger
func SetLogger(l load.Logger) {
	logger = l
	load.SetLogger(l)
}

// Inject dependencies
func Inject(dbas dataconnector.Storage, rs relation.Storage, ts table.Storage, ids id.Storage, dsfmap map[string]load.DataDestinationFactory, ri load.RowIterator) {
	dataconnectorStorage = dbas
	relStorage = rs
	tabStorage = ts
	idStorage = ids
	datadestinationFactories = dsfmap
	rowIterator = ri
}

// NewCommand implements the cli extract command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "load {<truncate>|<insert>} [Data Connector Name]",
		Short:   "Load data to a database with a loading mode (insert by default)",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s load truncate dstdatabase\n  %[1]s load dstdatabase", fullName),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return nil
			}
			if len(args) == 2 {
				if _, err := load.ParseMode(args[0]); err != nil {
					return err
				}
				return nil
			}
			return fmt.Errorf("accepts 1 or 2 args, received %d", len(args))
		},
		Run: func(cmd *cobra.Command, args []string) {
			var dcDestination = args[0]
			var mode, _ = load.ParseMode("insert")

			if len(args) == 2 {
				dcDestination = args[1]
				mode, _ = load.ParseMode(args[0])
			}

			datadestination, e1 := getDataDestination(dcDestination)
			if e1 != nil {
				fmt.Fprintln(err, e1.Error())
				os.Exit(1)
			}

			plan, e2 := getPlan()
			if e2 != nil {
				fmt.Fprintln(err, e2.Error())
				os.Exit(1)
			}
			logger.Debug(fmt.Sprintf("call Load with mode %s", mode))
			e3 := load.Load(rowIterator, datadestination, plan, mode)
			if e3 != nil {
				fmt.Fprintln(err, e3.Error())
				os.Exit(1)
			}
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}

func getDataDestination(dataconnectorName string) (load.DataDestination, *load.Error) {
	alias, e1 := dataconnector.Get(dataconnectorStorage, dataconnectorName)
	if e1 != nil {
		return nil, &load.Error{Description: e1.Error()}
	}
	if alias.ReadOnly {
		return nil, &load.Error{Description: fmt.Sprintf("'%s' is a read only dataconnector", alias.Name)}
	}

	u, e2 := dburl.Parse(alias.URL)
	if e2 != nil {
		return nil, &load.Error{Description: e2.Error()}
	}

	datadestinationFactory, ok := datadestinationFactories[u.Unaliased]
	if !ok {
		return nil, &load.Error{Description: "no datadestination found for database type"}
	}

	return datadestinationFactory.New(alias.URL), nil
}

func getPlan() (load.Plan, *load.Error) {
	id, err1 := idStorage.Read()
	if err1 != nil {
		return nil, &load.Error{Description: err1.Error()}
	}

	relations, err2 := relStorage.List()
	if err2 != nil {
		return nil, &load.Error{Description: err2.Error()}
	}

	tables, err3 := tabStorage.List()
	if err3 != nil {
		return nil, &load.Error{Description: err3.Error()}
	}

	rmap := map[string]relation.Relation{}
	for _, relation := range relations {
		rmap[relation.Name] = relation
	}

	tmap := map[string]table.Table{}
	for _, table := range tables {
		tmap[table.Name] = table
	}

	converter := idToLoadConverter{
		rmap:     rmap,
		tmap:     tmap,
		loadrmap: map[string]load.Relation{},
		loadtmap: map[string]load.Table{},
	}

	return converter.getPlan(id), nil
}

type idToLoadConverter struct {
	rmap map[string]relation.Relation
	tmap map[string]table.Table

	loadrmap map[string]load.Relation
	loadtmap map[string]load.Table
}

func (c idToLoadConverter) getTable(name string) load.Table {
	if loadtable, ok := c.loadtmap[name]; ok {
		return loadtable
	}

	table, ok := c.tmap[name]
	if !ok {
		logger.Error(fmt.Sprintf("missing table %v in tables.yaml", name))
		return load.NewTable(name, "")
	}

	logger.Trace(fmt.Sprintf("building table %v", table))

	return load.NewTable(table.Name, table.Keys[0]) // TODO : support multivalued primary keys
}

func (c idToLoadConverter) getRelation(name string) load.Relation {
	if loadrelation, ok := c.loadrmap[name]; ok {
		return loadrelation
	}

	relation, ok := c.rmap[name]
	if !ok {
		logger.Error(fmt.Sprintf("missing relation %v in relations.yaml", name))
		return load.NewRelation(name, nil, nil)
	}

	logger.Trace(fmt.Sprintf("building relation %v", relation))

	return load.NewRelation(
		relation.Name,
		c.getTable(relation.Parent.Name),
		c.getTable(relation.Child.Name),
	)
}

func (c idToLoadConverter) getPlan(id id.IngressDescriptor) load.Plan {
	relations := []load.Relation{}

	for idx := uint(0); idx < id.Relations().Len(); idx++ {
		rel := id.Relations().Relation(idx)
		relations = append(relations, c.getRelation(rel.Name()))
	}

	return load.NewPlan(c.getTable(id.StartTable().Name()), relations)
}
