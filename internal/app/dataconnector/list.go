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

package dataconnector

import (
	"fmt"
	"os"

	"github.com/cgi-fr/lino/pkg/dataconnector"
	"github.com/spf13/cobra"
)

// newListCommand implements the cli dataconnector list command
func newListCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List database aliases",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s dataconnector list", fullName),
		Args:    cobra.NoArgs,
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			list, e := dataconnector.List(storage)
			if e != nil {
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}

			for _, m := range list {
				fmt.Fprintln(out, m)
			}
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
