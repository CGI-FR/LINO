// Copyright (C) 2021 CGI France
//
// This file is part of LINO.
//
// LINO is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// LINO is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with LINO.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"

	"github.com/cgi-fr/lino/internal/app/dataconnector"
	"github.com/cgi-fr/lino/internal/app/http"
	"github.com/cgi-fr/lino/internal/app/id"
	"github.com/cgi-fr/lino/internal/app/pull"
	"github.com/cgi-fr/lino/internal/app/push"
	"github.com/cgi-fr/lino/internal/app/relation"
	"github.com/cgi-fr/lino/internal/app/table"
	"github.com/spf13/cobra"
)

// Provisioned by ldflags
// nolint: gochecknoglobals
var (
	version   string
	commit    string
	buildDate string
	builtBy   string

	logger Logger

	// global flags
	loglevel string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lino [action]",
	Short: "Command line tools for managing tests data",
	Long:  `Lino is a simple ETL (Extract Transform Load) tools to manage tests datas. The lino command line tool pull test data from a relational database to create a smallest production-like database.`,
	Example: `  lino dataconnector add source --read-only postgresql://postgres@localhost:5432/postgres?sslmode=disable
  lino dc add target postgresql://postgres@localhost:5433/postgres?sslmode=disable
  lino dc list
  lino table extract source
  lino relation extract source
  lino id create customer
  lino id display-plan
  lino id show-graph
  lino pull source --limit 10 > customers.jsonl
  lino push target < customers.jsonl`,
	Version: fmt.Sprintf(`%v (commit=%v date=%v by=%v)
Copyright (C) 2021 CGI France
License GPLv3: GNU GPL version 3 <https://gnu.org/licenses/gpl.html>.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.`, version, commit, buildDate, builtBy),
}

func main() {
	// CPU profiling code starts here
	/* 	f, _ := os.Create("lino.cpu.prof")
	   	defer f.Close()
	   	pprof.StartCPUProfile(f)
	   	defer pprof.StopCPUProfile() */
	// CPU profiling code ends here

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// global flags
	rootCmd.PersistentFlags().StringVarP(&loglevel, "verbosity", "v", "none", "set level of log verbosity : none (0), error (1), warn (2), info (3), debug (4), trace (5)")
	rootCmd.AddCommand(dataconnector.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(table.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(relation.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(id.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(pull.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(push.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(http.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
}

func initConfig() {
	switch loglevel {
	case "trace", "5":
		logger = NewLogger(os.Stderr, os.Stderr, os.Stderr, os.Stderr, os.Stderr)
		logger.Trace("Logger level set to trace")
	case "debug", "4":
		logger = NewLogger(nil, os.Stderr, os.Stderr, os.Stderr, os.Stderr)
		logger.Debug("Logger level set to debug")
	case "info", "3":
		logger = NewLogger(nil, nil, os.Stderr, os.Stderr, os.Stderr)
		logger.Info("Logger level set to info")
	case "warn", "2":
		logger = NewLogger(nil, nil, nil, os.Stderr, os.Stderr)
		logger.Warn("Logger level set to warn")
	case "error", "1":
		logger = NewLogger(nil, nil, nil, nil, os.Stderr)
		logger.Error("Logger level set to error")
	default:
		logger = NewLogger(nil, nil, nil, nil, nil)
	}

	dataconnector.SetLogger(logger)
	id.SetLogger(logger)
	pull.SetLogger(logger)
	push.SetLogger(logger)

	dataconnector.Inject(dataconnectorStorage(), dataPingerFactory())
	relation.Inject(dataconnectorStorage(), relationStorage(), relationExtractorFactory())
	table.Inject(dataconnectorStorage(), tableStorage(), tableExtractorFactory())
	id.Inject(idStorage(), relationStorage(), idExporter(), idJSONStorage(*os.Stdout))
	pull.Inject(dataconnectorStorage(), relationStorage(), tableStorage(), idStorageFactory(), pullDataSourceFactory(), pullRowExporterFactory(), pullRowReaderFactory(), traceListner(os.Stderr))
	push.Inject(dataconnectorStorage(), relationStorage(), tableStorage(), idStorageFactory(), pushDataDestinationFactory(), pushRowIteratorFactory(), pushRowExporterFactory())
}
