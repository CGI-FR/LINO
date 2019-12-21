package id

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newExportCommand implements the cli id export command
func newExportCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export",
		Short:   "Export content of ingress descriptor in JSON format to stdout",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s id export", fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			id, e := idStorage.Read()
			if e != nil {
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}
			e = idJSONExporter.Store(id)
			if e != nil {
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
