package id

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	infra "makeit.imfr.cgi.com/lino/internal/infra/id"
	"makeit.imfr.cgi.com/lino/pkg/id"
)

// newCreateCommand implements the cli id create command
func newCreateCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create [Start table]",
		Short:   "Create ingress descriptor",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s id create public.customer", fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			table := args[0]

			relations, e1 := relStorage.List()
			if e1 != nil {
				fmt.Fprintln(err, e1.Description)
				os.Exit(1)
			}

			reader := infra.NewRelationReader(relations)

			e := id.Create(table, reader, idStorage)
			if e != nil {
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}

			fmt.Fprintln(out, "successfully created ingress descriptor")
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
