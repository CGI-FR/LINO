package table

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
	"makeit.imfr.cgi.com/lino/pkg/table"
)

var dataconnectorStorage dataconnector.Storage
var tableStorage table.Storage
var tableExtractorFactories map[string]table.ExtractorFactory

// Inject dependencies
func Inject(dbas dataconnector.Storage, rs table.Storage, exmap map[string]table.ExtractorFactory) {
	dataconnectorStorage = dbas
	tableStorage = rs
	tableExtractorFactories = exmap
}

// NewCommand implements the cli dataconnector command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "table {extract} [arguments ...]",
		Short:   "Manage tables",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s table extract mydatabase", fullName),
		Aliases: []string{"tab"},
	}
	cmd.AddCommand(newExtractCommand(fullName, err, out, in))
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
