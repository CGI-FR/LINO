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
	var flagURLFromEnv string

	cmd := &cobra.Command{
		Use:     "add [Name] [URL]",
		Short:   "Add database aliases",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s dataconnector add mydatabase postgresql://postgres:sakila@localhost:5432/postgres?sslmode=disable", fullName),
		Args: func(cmd *cobra.Command, args []string) error {
			if flagURLFromEnv != "" {
				if flagAskPassword {
					return fmt.Errorf("cannot use --password with --url-from-env, use --password-from-env")
				}
				return cobra.ExactArgs(1)(cmd, args)
			}
			return cobra.ExactArgs(2)(cmd, args)
		},
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			var u *dburl.URL
			username := ""
			password := ""
			if len(args) == 2 {
				var e2 error
				u, e2 = dburl.Parse(args[1])
				if e2 != nil {
					fmt.Fprintln(err, e2.Error()) //nolint:errcheck
					os.Exit(3)
				}
				var isset bool
				password, isset = u.User.Password()
				if isset {
					fmt.Fprintln(err, "warn: password should not be included in URI, use --password-from-env or --password") //nolint:errcheck
					u.User = url.User(u.User.Username())
				}
				username = u.User.Username()
			}
			if flagUserValue != "" {
				username = flagUserValue
				if u != nil {
					u.User = url.User(username)
				}
			}

			if flagAskPassword {
				switch {
				case username != "":
					password = askPassword()
				case flagUserFromEnv != "":
					fmt.Fprintln(err, "error: cannot use --password with --user-from-env, use --password-from-env or specify a username") //nolint:errcheck
					os.Exit(1)
				default:
					fmt.Fprintln(err, "error: cannot use --password with empty username, use --password-from-env or specify a username") //nolint:errcheck
					os.Exit(1)
				}
			}

			if password != "" && u != nil {
				e3 := urlbuilder.StorePassword(u, password, err)
				if e3 != nil {
					fmt.Fprintln(err, e3.Error()) //nolint:errcheck
					os.Exit(3)
				}
			}
			connectionURL := ""
			if u != nil {
				connectionURL = u.URL.String()
			}

			alias := dataconnector.DataConnector{
				Name: name,
				URL: dataconnector.ValueHolder{
					Value:        connectionURL,
					ValueFromEnv: flagURLFromEnv,
				},
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
				fmt.Fprintln(err, e.Description) //nolint:errcheck
				os.Exit(1)
			}

			fmt.Fprintf(out, "successfully added dataconnector") //nolint:errcheck
			fmt.Fprintln(out)                                    //nolint:errcheck
		},
	}
	cmd.Flags().BoolVarP(&flagReadonly, "read-only", "r", false, "Write protection flag that prevents modification")
	cmd.Flags().StringVarP(&flagSchema, "schema", "s", "", "Default schema to use with that dataconnector")
	cmd.Flags().StringVar(&flagURLFromEnv, "url-from-env", "", "Name of environment variable containing URL")
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
		os.Stdout.Write([]byte("enter password: ")) //nolint:errcheck,gosec
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		os.Stdout.Write([]byte("\n")) //nolint:errcheck,gosec
		if err != nil {
			os.Exit(1)
		}
		return string(bytePassword)
	}
	return ""
}
