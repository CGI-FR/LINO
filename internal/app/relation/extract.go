package relation

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
	"makeit.imfr.cgi.com/lino/pkg/relation"
)

var schema string

// newExtractCommand implements the cli relation extract command
func newExtractCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extract [DB Alias Name]",
		Short:   "Extract relations from database",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s relation extract mydatabase", fullName),
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

			factory, ok := relationExtractorFactories[u.Unaliased]
			if !ok {
				fmt.Fprintln(err, "no extractor found for database type")
				os.Exit(1)
			}

			extractor := factory.New(alias.URL, schema)

			e2 := relation.Extract(extractor, relationStorage)
			if e2 != nil {
				fmt.Fprintln(err, e2.Description)
				os.Exit(1)
			}

			relations, e2 := relationStorage.List()
			if e2 != nil {
				fmt.Fprintln(err, e2.Description)
				os.Exit(1)
			}

			fmt.Fprintf(out, "lino finds %v relations from constraints", len(relations))
		},
	}
	cmd.Flags().StringVarP(&schema, "schema", "s", "", "specify the schema to use")
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
