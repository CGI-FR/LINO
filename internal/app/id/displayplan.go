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

// newDisplayPlanCommand implements the cli id display-plan command
func newDisplayPlanCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "display-plan",
		Short:   "Show ingress descriptor steps",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s id display-plan", fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			result, e := id.GetPullerPlan(idStorage)
			if e != nil {
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}

			for i := uint(0); i < result.Len(); i++ {
				fmt.Fprintln(out, result.Step(i))
			}
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
