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

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/pkg/id"
)

// newSetParentLookupCommand implements the cli id set-parent-lookup command
func newSetParentLookupCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-parent-lookup [relation] [true|false]",
		Short:   "set parent lookup flag for relation [relation] in ingress descriptor",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s id set-parent-lookup public.store", fullName),
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			relation := args[0]
			flag := args[1] == "true"
			if !flag && args[1] != "false" {
				fmt.Fprintln(err, "flag must be 'true' or 'false'")
				os.Exit(1)
			}

			e := id.SetParentLookup(relation, flag, idStorage)
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
