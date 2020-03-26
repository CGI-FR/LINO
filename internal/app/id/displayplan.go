package id

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/pkg/id"
)

// newDisplayPlanCommand implements the cli id display-plan command
func newDisplayPlanCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "display-plan",
		Short:   "Show ingress descriptor steps",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s id display-plan", fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			result, e := id.GetPullionPlan(idStorage)
			if e != nil {
				fmt.Fprintln(err, e.Description)
				os.Exit(1)
			}

			for i := uint(0); i < result.Len(); i++ {
				fmt.Fprintln(out, result.Step(i))
			}
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
