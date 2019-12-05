package table

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
	"makeit.imfr.cgi.com/lino/pkg/table"
)

// newExtractCommand implements the cli relation extract command
func newExtractCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extract [DB Alias Name]",
		Short:   "Extract tables metadatas from database",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s table extract mydatabase", fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			alias, e1 := dataconnector.Get(dataconnectorStorage, args[0])
			if e1 != nil {
				fmt.Fprintln(err, e1.Description)
				os.Exit(1)
			}

			if alias == nil {
				fmt.Fprintln(err, "no dataconnector named "+args[0])
				os.Exit(1)
			}

			u, e := dburl.Parse(alias.URL)
			if e != nil {
				fmt.Fprintln(err, e)
				os.Exit(1)
			}

			factory, ok := tableExtractorFactories[u.Unaliased]
			if !ok {
				fmt.Fprintln(err, "no extractor found for database type")
				os.Exit(1)
			}

			extractor := factory.New(alias.URL)

			e2 := table.Extract(extractor, tableStorage)
			if e2 != nil {
				fmt.Fprintln(err, e2.Description)
				os.Exit(1)
			}

			tables, e2 := tableStorage.List()
			if e2 != nil {
				fmt.Fprintln(err, e2.Description)
				os.Exit(1)
			}

			fmt.Fprintf(out, "lino finds %v table(s)", len(tables))
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
