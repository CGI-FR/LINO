package relation

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
	"makeit.imfr.cgi.com/lino/pkg/relation"
)

var dataconnectorStorage dataconnector.Storage
var relationStorage relation.Storage
var relationExtractorFactories map[string]relation.ExtractorFactory

// Inject dependencies
func Inject(dbas dataconnector.Storage, rs relation.Storage, exmap map[string]relation.ExtractorFactory) {
	dataconnectorStorage = dbas
	relationStorage = rs
	relationExtractorFactories = exmap
}

// NewCommand implements the cli dataconnector command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "relation {extract} [arguments ...]",
		Short:   "Manage relations",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s relation extract mydatabase", fullName),
		Aliases: []string{"rel"},
	}
	cmd.AddCommand(newExtractCommand(fullName, err, out, in))
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
