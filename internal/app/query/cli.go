package query

import (
	"fmt"
	"os"

	"github.com/cgi-fr/lino/internal/app/urlbuilder"
	infra "github.com/cgi-fr/lino/internal/infra/query"
	"github.com/cgi-fr/lino/pkg/dataconnector"
	"github.com/cgi-fr/lino/pkg/query"
	"github.com/spf13/cobra"
)

var (
	dataconnectorStorage dataconnector.Storage
	dataSourceFactories  map[string]infra.DataSourceFactory
)

// Inject dependencies
func Inject(
	dbas dataconnector.Storage,
	dsfs map[string]infra.DataSourceFactory,
) {
	dataconnectorStorage = dbas
	dataSourceFactories = dsfs
}

// NewCommand implements the cli analyse command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query",
		Short:   "Execute direct query",
		Example: fmt.Sprintf("  %[1]s", fullName),
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if er := execute(cmd, args[0], args[1]); er != nil {
				fmt.Fprintln(err, er.Error())
				os.Exit(1)
			}
		},
	}

	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)

	return cmd
}

func execute(cmd *cobra.Command, dataconnectorName string, querystr string) error {
	alias, e1 := dataconnector.Get(dataconnectorStorage, dataconnectorName)
	if e1 != nil {
		return e1
	}

	if alias == nil {
		return fmt.Errorf("Data Connector %s not found", dataconnectorName)
	}

	u := urlbuilder.BuildURL(alias, cmd.OutOrStdout())

	dataSourceFactory, ok := dataSourceFactories[u.UnaliasedDriver]
	if !ok {
		return fmt.Errorf("no extractor found for database type")
	}

	driver := query.NewDriver(dataSourceFactory.New(u.URL.String()), nil)

	if err := driver.Open(); err != nil {
		return fmt.Errorf("%w", err)
	}

	defer driver.Close()

	if err := driver.Execute(querystr); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
