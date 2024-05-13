package query

import (
	"fmt"
	"os"

	"github.com/cgi-fr/lino/pkg/query"
	"github.com/spf13/cobra"
)

// NewCommand implements the cli analyse command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query",
		Short:   "Execute direct query",
		Example: fmt.Sprintf("  %[1]s", fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			execute(args[0])
		},
	}

	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)

	return cmd
}

func execute(querystr string) {
	driver := query.NewDriver(nil, nil)

	driver.Open()
	driver.Execute(querystr)
	driver.Close()
}
