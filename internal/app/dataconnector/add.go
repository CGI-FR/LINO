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
	"net/url"
	"os"

	"github.com/cgi-fr/lino/internal/app/urlbuilder"
	"github.com/cgi-fr/lino/pkg/dataconnector"
	"github.com/spf13/cobra"
	"github.com/xo/dburl"
	"golang.org/x/term"
)

// newAddCommand implements the cli dataconnector add command
func newAddCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	var flagReadonly bool
	var flagSchema string
	var flagAskPassword bool
	var flagPasswordFromEnv string
	var flagUserFromEnv string
	var flagUserValue string

	cmd := &cobra.Command{
		Use:     "add [Name] [URL]",
		Short:   "Add database aliases",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s dataconnector add mydatabase postgresql://postgres:sakila@localhost:5432/postgres?sslmode=disable", fullName),
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			urlin := args[1]

			u, e2 := dburl.Parse(urlin)
			if e2 != nil {
				fmt.Fprintln(err, e2.Error())
				os.Exit(3)
			}

			password, isset := u.User.Password()
			if isset {
				fmt.Fprintln(err, "warn: password should not be included in URI, use --password-from-env or --password")
				u.User = url.User(u.User.Username())
			}

			username := u.User.Username()
			if flagUserValue != "" {
				username = flagUserValue
				u.User = url.User(username)
			}

			if flagAskPassword {
				switch {
				case username != "":
					password = askPassword()
				case flagUserFromEnv != "":
					fmt.Fprintln(err, "error: cannot use --password with --user-from-env, use --password-from-env or specify a username")
					os.Exit(1)
				default:
					fmt.Fprintln(err, "error: cannot use --password with empty username, use --password-from-env or specify a username")
					os.Exit(1)
				}
			}

			if password != "" {
				e3 := urlbuilder.StorePassword(u, password, err)
				if e3 != nil {
					fmt.Fprintln(err, e3.Error())
					os.Exit(3)
				}
			}

			alias := dataconnector.DataConnector{
				Name:     name,
				URL:      u.URL.String(),
				ReadOnly: flagReadonly,
				Schema:   flagSchema,
				User: dataconnector.ValueHolder{
					Value:        "",
					ValueFromEnv: flagUserFromEnv,
				},
				Password: dataconnector.ValueHolder{
					Value:        "",
					ValueFromEnv: flagPasswordFromEnv,
				},
			}

			e := dataconnector.Add(storage, &alias)
			if e != nil {
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}

			fmt.Fprintf(out, "successfully added dataconnector")
			fmt.Fprintln(out)
		},
	}
	cmd.Flags().BoolVarP(&flagReadonly, "read-only", "r", false, "Write protection flag that prevents modification")
	cmd.Flags().StringVarP(&flagSchema, "schema", "s", "", "Default schema to use with that dataconnector")
	cmd.Flags().StringVarP(&flagPasswordFromEnv, "password-from-env", "P", "", "Name of environment variable containing password")
	cmd.Flags().StringVarP(&flagUserFromEnv, "user-from-env", "U", "", "Name of environment variable containing username")
	cmd.Flags().StringVarP(&flagUserValue, "user", "u", "", "Username to connect")
	cmd.Flags().BoolVarP(&flagAskPassword, "password", "p", false, "Ask password from terminal prompt")
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}

func askPassword() string {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		os.Stdout.Write([]byte("enter password: "))
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		os.Stdout.Write([]byte("\n"))
		if err != nil {
			os.Exit(1)
		}
		return string(bytePassword)
	}
	return ""
}
