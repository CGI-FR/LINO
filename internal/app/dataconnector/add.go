package dataconnector

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
)

// newAddCommand implements the cli dataconnector add command
func newAddCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	var schema string
	var passwordFromEnv string
	var userFromEnv string
	var user string

	cmd := &cobra.Command{
		Use:     "add [Name] [URL]",
		Short:   "Add database aliases",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s dataconnector add mydatabase postgresql://postgres:sakila@localhost:5432/postgres?sslmode=disable", fullName),
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			url := args[1]

			u, e2 := dburl.Parse(url)
			if e2 != nil {
				fmt.Fprintln(err, e2.Error())
				os.Exit(3)
			}

			if _, isset := u.User.Password(); isset {
				fmt.Fprintln(err, "warn: password should not be included in URI, use --password-from-env")
			}

			alias := dataconnector.DataConnector{
				Name:     name,
				URL:      url,
				ReadOnly: readonly,
				Schema:   schema,
				User: dataconnector.ValueHolder{
					Value:        user,
					ValueFromEnv: userFromEnv,
				},
				Password: dataconnector.ValueHolder{
					Value:        "",
					ValueFromEnv: passwordFromEnv,
				},
			}

			e := dataconnector.Add(storage, &alias)
			if e != nil {
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}

			fmt.Fprintf(out, "successfully added dataconnector %v", alias)
			fmt.Fprintln(out)
		},
	}
	cmd.Flags().BoolVarP(&readonly, "read-only", "r", false, "Write protection flag that prevents modification")
	cmd.Flags().StringVarP(&schema, "schema", "s", "", "Default schema to use with that dataconnector")
	cmd.Flags().StringVarP(&passwordFromEnv, "password-from-env", "P", "", "Name of environment variable containing password")
	cmd.Flags().StringVarP(&userFromEnv, "user-from-env", "U", "", "Name of environment variable containing username")
	cmd.Flags().StringVarP(&user, "user", "u", "", "Username to connect")
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
