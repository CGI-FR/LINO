package dataconnector

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
)

// newAddCommand implements the cli dataconnector add command
func newAddCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [Name] [URL]",
		Short:   "Add database aliases",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s dataconnector add mydatabase postgresql://postgres:sakila@localhost:5432/postgres?sslmode=disable", fullName),
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			url := args[1]

			alias := dataconnector.DataConnector{
				Name:     name,
				URL:      url,
				ReadOnly: readonly,
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
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
