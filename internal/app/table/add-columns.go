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

package table

import (
	"fmt"
	"os"

	"github.com/cgi-fr/lino/pkg/table"
	"github.com/spf13/cobra"
)

// newAddColumnCommand implements the cli table add-column command
func newAddColumnCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	// local flags
	var exportType, importType string

	cmd := &cobra.Command{
		Use:     "add-column [Table Name] [Column Name]",
		Short:   "Add column definition in tables.yaml file",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s add-column public.actor first_name", fullName),
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			tableName := args[0]
			columnName := args[1]

			_, e1 := table.AddOrUpdateColumn(tableStorage, tableName, columnName, exportType, importType)
			if e1 != nil {
				fmt.Fprintln(err, e1.Description)
				os.Exit(1)
			}

			fmt.Fprintf(out, "successfully added column %v to %v table\n", columnName, tableName)
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	cmd.Flags().StringVarP(&exportType, "export", "e", "", "export type for the column")
	cmd.Flags().StringVarP(&importType, "import", "i", "", "import type for the column")
	return cmd
}
