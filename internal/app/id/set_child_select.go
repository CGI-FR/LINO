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

package id

import (
	"fmt"
	"os"

	"github.com/cgi-fr/lino/pkg/id"
	"github.com/spf13/cobra"
)

// newSetChildSelectCommand implements the cli id set-child-select command
func newSetChildSelectCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-child-select [relation] [column1] [column2] [column3] ...",
		Short:   "set child select attribut for relation [relation] in ingress descriptor",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s id set-child-select public.store store_id name", fullName),
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			relation := args[0]
			selectColumns := args[1:]

			e := id.SetChildSelect(relation, selectColumns, idStorageFactory(ingressDescriptor))
			if e != nil {
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}

			fmt.Fprintf(out, "successfully update relation %s in ingress descriptor\n", relation)
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
