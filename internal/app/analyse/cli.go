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

package analyse

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cgi-fr/lino/internal/app/urlbuilder"
	infra "github.com/cgi-fr/lino/internal/infra/analyse"
	"github.com/cgi-fr/lino/pkg/analyse"
	"github.com/cgi-fr/lino/pkg/dataconnector"
	"github.com/cgi-fr/lino/pkg/table"
	"github.com/spf13/cobra"
)

var (
	tableStorage         table.Storage
	dataconnectorStorage dataconnector.Storage
	extractorFactories   map[string]infra.SQLExtractorFactory
)

const DefaultSampleSize = uint(5)

// Inject dependencies
func Inject(
	ts table.Storage,
	dbas dataconnector.Storage,
	dsf map[string]infra.SQLExtractorFactory,
) {
	tableStorage = ts
	dataconnectorStorage = dbas
	extractorFactories = dsf
}

// NewCommand implements the cli analyse command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	// local flags
	var distinct bool
	var limit uint
	var tables []string
	var wheres map[string]string
	var exclude map[string]string
	var excludePks bool
	var sampleSize uint

	cmd := &cobra.Command{
		Use:     "analyse",
		Short:   "Analyse database content",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s", fullName),
		Aliases: []string{"rimo"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dataConnector := args[0]
			dataSource, primaryKeys, e0 := getDatasource(dataConnector)
			if e0 != nil {
				fmt.Fprintln(err, e0.Error())
				os.Exit(1)
			}

			excludedColumns := splitColumns(exclude)
			if excludePks {
				excludedColumns = mergeColumns(excludedColumns, primaryKeys)
			}

			extractor, e1 := getExtractor(dataConnector, out)
			if e1 != nil {
				fmt.Fprintln(err, e1.Error())
				os.Exit(1)
			}

			writer := getWriter(out)

			driver := analyse.NewDriver(dataSource, extractor, writer,
				analyse.Config{
					SampleSize:     sampleSize,
					Distinct:       distinct,
					Limit:          limit,
					Tables:         tables,
					Wheres:         wheres,
					ExcludeColumns: excludedColumns,
				},
			)
			if e2 := driver.Analyse(); e2 != nil {
				fmt.Fprintf(err, "analyse failed '%s'", dataConnector)
				fmt.Fprintln(err)
				os.Exit(5)
			}
		},
	}

	cmd.Flags().UintVarP(&limit, "limit", "l", 0, "limit the number of results (0 = no limit)")
	cmd.Flags().BoolVarP(&distinct, "distinct", "D", false, "count distinct values")
	cmd.Flags().StringArrayVarP(&tables, "table", "t", []string{}, "specify tables to analyse")
	cmd.Flags().StringToStringVarP(&wheres, "where", "w", map[string]string{}, "where clauses by table")
	cmd.Flags().StringToStringVarP(&exclude, "exclude", "e", map[string]string{}, "specify columns to exclude by table")
	cmd.Flags().BoolVarP(&excludePks, "exclude-pk", "x", false, "exclude primary keys of each tables")
	cmd.Flags().UintVar(&sampleSize, "sample-size", DefaultSampleSize, "number of sample value to collect")

	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}

func getWriter(out io.Writer) analyse.Writer {
	return infra.NewStdWriter(out)
}

func getExtractor(dataconnectorName string, out io.Writer) (analyse.ExtractorFactory, error) {
	alias, e1 := dataconnector.Get(dataconnectorStorage, dataconnectorName)
	if e1 != nil {
		return nil, e1
	}
	if alias == nil {
		return nil, fmt.Errorf("Data Connector %s not found", dataconnectorName)
	}

	u := urlbuilder.BuildURL(alias, out)

	datasourceFactory, ok := extractorFactories[u.UnaliasedDriver]
	if !ok {
		return nil, fmt.Errorf("no extractor found for database type")
	}

	return datasourceFactory.New(u.URL.String(), alias.Schema), nil
}

func getDatasource(dataconnectorName string) (analyse.DataSource, map[string][]string, error) {
	primaryKeys := map[string][]string{}
	result := map[string][]string{}
	tables, err := tableStorage.List()
	if err != nil {
		return nil, nil, err
	}

	for _, table := range tables {
		columns := []string{}
		for _, column := range table.Columns {
			columns = append(columns, column.Name)
		}
		result[table.Name] = columns

		primaryKeys[table.Name] = table.Keys
	}

	return infra.NewMapDataSource(dataconnectorName, result), primaryKeys, nil
}

func splitColumns(exclude map[string]string) map[string][]string {
	result := make(map[string][]string, len(exclude))
	for table, columns := range exclude {
		result[table] = strings.Split(columns, ",")
	}
	return result
}

func mergeColumns(excludedColumns map[string][]string, primaryKeys map[string][]string) map[string][]string {
	result := make(map[string][]string, len(excludedColumns)+len(primaryKeys))
	for table, columns := range primaryKeys {
		result[table] = columns
	}
	for table, columns := range excludedColumns {
		if existing, ok := result[table]; ok {
			result[table] = append(existing, columns...)
		} else {
			result[table] = columns
		}
	}
	return result
}
