package dataconnector

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/xo/dburl"
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
				fmt.Fprintln(err, e.Description)
				os.Exit(2)
			}
			if dc == nil {
				fmt.Fprintf(err, "no dataconnector for '%s'", args[0])
				fmt.Fprintln(err)
				os.Exit(5)
			}
			u, e2 := dburl.Parse(dc.URL)
			if e2 != nil {
				fmt.Fprintln(err, e2.Error())
				os.Exit(3)
			}
			// get user from env
			if dc.User.ValueFromEnv != "" {
				userFromEnv := os.Getenv(dc.User.ValueFromEnv)
				if userFromEnv == "" {
					fmt.Fprintf(err, "warn: missing environment variable %s", dc.User.ValueFromEnv)
					fmt.Fprintln(err)
				} else {
					u.User = url.User(userFromEnv)
					logger.Debug(fmt.Sprintf("ping user = %s", u.User))
				}
			} else if dc.User.Value != "" {
				// set user from dc
				u.User = url.User(dc.User.Value)
			}
			// get password from env
			if dc.Password.ValueFromEnv != "" {
				passwordFromEnv := os.Getenv(dc.Password.ValueFromEnv)
				if passwordFromEnv == "" {
					fmt.Fprintf(err, "warn: missing environment variable %s", dc.Password.ValueFromEnv)
					fmt.Fprintln(err)
				} else {
					u.User = url.UserPassword(u.User.Username(), passwordFromEnv)
				}
			}
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
