// Copyright (C) 2022 CGI France
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

package table

import (
	"fmt"
	"os"

	"github.com/cgi-fr/lino/internal/app/urlbuilder"
	"github.com/cgi-fr/lino/pkg/dataconnector"

	"github.com/cgi-fr/lino/pkg/table"
	"github.com/spf13/cobra"
)

// newCountCommand implements the cli table count command
func newCountCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "count [Data Connector]",
		Short:   "Count lines in the tables using dataconnector [Data Connector]",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s table count mydatabase", fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			alias, e1 := dataconnector.Get(dataconnectorStorage, args[0])
			if e1 != nil {
				fmt.Fprintln(err, e1.Description)
				os.Exit(1)
			}

			if alias == nil {
				fmt.Fprintln(err, "no dataconnector named "+args[0])
				os.Exit(1)
			}

			u := urlbuilder.BuildURL(alias, err)

			factory, ok := tableExtractorFactories[u.UnaliasedDriver]
			if !ok {
				fmt.Fprintln(err, "no extractor found for database type")
				os.Exit(1)
			}

			extractor := factory.New(u.URL.String(), alias.Schema, false)

			result, e2 := table.Count(tableStorage, extractor)
			if e2 != nil {
				fmt.Fprintln(err, e2.Description)
				os.Exit(1)
			}
			for _, tableCount := range result {
				fmt.Printf("%s: %d\n", tableCount.Table.Name, tableCount.Count)
			}
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
