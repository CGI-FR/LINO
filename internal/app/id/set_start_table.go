package id

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/pkg/id"
)

// newSetStartTableCommand implements the cli id set-start-table command
func newSetStartTableCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-start-table [Start table]",
		Short:   "set new start table ingress descriptor",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s id set-start-table public.store", fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			table := args[0]

			e := id.SetStartTable(id.NewTable(table), idStorage)
			if e != nil {
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}

			fmt.Fprintln(out, "successfully update start table ingress descriptor")
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
