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
	"os"

	"github.com/cgi-fr/rimo/pkg/rimo"
	"github.com/spf13/cobra"
)

var (
	reader rimo.Reader
	writer rimo.Writer
)

// Inject dependencies
func Inject(r rimo.Reader, w rimo.Writer) {
	reader = r
	writer = w
}

// NewCommand implements the cli analyse command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "analyse",
		Short:   "Analyse database content",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s", fullName),
		Aliases: []string{"rimo"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dataConnector := args[0]
			e1 := rimo.AnalyseBase(reader, writer)
			if e1 != nil {
				fmt.Fprintf(err, "analyse faield '%s'", dataConnector)
				fmt.Fprintln(err)
				os.Exit(5)
			}
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
