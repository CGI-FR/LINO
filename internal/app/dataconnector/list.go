package dataconnector

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
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
