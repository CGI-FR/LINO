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

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/internal/app/urlbuilder"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
)

// newListCommand implements the cli dataconnector list command
func newPingCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ping",
		Short:   "Ping database",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s dataconnector ping source", fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dc, e := dataconnector.Get(storage, args[0])
			if e != nil {
				logger.Error(fmt.Sprintf(e.Description))
				fmt.Fprintln(err, e.Description)
				os.Exit(2)
			}
			if dc == nil {
				fmt.Fprintf(err, "no dataconnector for '%s'", args[0])
				fmt.Fprintln(err)
				os.Exit(5)
			}
			u := urlbuilder.BuildURL(dc, err)
			dataPingerFactory, ok := dataPingerFactory[u.Unaliased]
			if !ok {
				fmt.Fprintln(err, "no datadestination found for database type")
				os.Exit(4)
			}
			pinger := dataPingerFactory.New(u.URL.String())
			e = pinger.Ping()
			if e != nil {
				fmt.Fprintln(out, "ping failed")
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}

			fmt.Fprintln(out, "ping success")
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
