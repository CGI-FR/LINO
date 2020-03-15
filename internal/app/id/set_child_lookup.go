package id

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/pkg/id"
)

// newSetChildLookupCommand implements the cli id set-child-lookup command
func newSetChildLookupCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-child-lookup [relation] [true|false]",
		Short:   "set child lookup flag for relation [relation] in ingress descriptor",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s id set-child-lookup public.store", fullName),
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			relation := args[0]
			flag := args[1] == "true"
			if !flag && args[1] != "false" {
				fmt.Fprintln(err, "flag must be 'true' or 'false'")
				os.Exit(1)
			}

			e := id.SetChildLookup(relation, flag, idStorage)
			if e != nil {
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}

			fmt.Fprintf(out, "successfully update relation %s in ingress descriptor\n", relation)
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
