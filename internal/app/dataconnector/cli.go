package dataconnector

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
)

var storage dataconnector.Storage
var dataPingerFactory map[string]dataconnector.DataPingerFactory

// local flags
var readonly bool

// Inject dependencies
func Inject(dbas dataconnector.Storage, dpf map[string]dataconnector.DataPingerFactory) {
	storage = dbas
	dataPingerFactory = dpf
}

// NewCommand implements the cli dataconnector command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dataconnector {add,list} [arguments ...]",
		Short:   "Manage database aliases",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s dataconnector add mydatabase postgresql://postgres:sakila@localhost:5432/postgres?sslmode=disable", fullName),
		Aliases: []string{"dc"},
	}
	cmd.AddCommand(newAddCommand(fullName, err, out, in))
	cmd.AddCommand(newListCommand(fullName, err, out, in))
	cmd.AddCommand(newPingCommand(fullName, err, out, in))
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
