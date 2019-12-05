package id

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/pkg/id"
)

// newShowGraphCommand implements the cli id show-graph command
func newShowGraphCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-graph",
		Short:   "Show ingress descriptor graph",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s id show-graph", fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			e := id.Export(idStorage, idExporter)
			if e != nil {
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}

			fmt.Fprintln(out, "success")
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
