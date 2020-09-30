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
