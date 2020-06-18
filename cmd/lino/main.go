package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makeit.imfr.cgi.com/lino/internal/app/dataconnector"
	"makeit.imfr.cgi.com/lino/internal/app/http"
	"makeit.imfr.cgi.com/lino/internal/app/id"
	"makeit.imfr.cgi.com/lino/internal/app/pull"
	"makeit.imfr.cgi.com/lino/internal/app/push"
	"makeit.imfr.cgi.com/lino/internal/app/relation"
	"makeit.imfr.cgi.com/lino/internal/app/table"
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
	Long:  `Lino is a simple ETL (Extract Transform Push) tools to manage tests datas. The lino command line tool pull test data from a relational database to create a smallest production-like database.`,
	Example: `  lino dataconnector add mydatabase postgresql://postgres:sakila@localhost:5432/postgres?sslmode=disable
  lino db list
  lino table extract mydatabase
  lino relation extract mydatabase
  lino id create [Table Name]
  lino id display-plan
  lino id show-graph
  lino pull mydatabase --limit 10 > customers.jsonl
  lino push customer --input customer.json --jdbc jdbc:oracle:thin:scott/tiger@target:1721:xe`,
	Version: fmt.Sprintf("%v (commit=%v date=%v by=%v)", version, commit, buildDate, builtBy),
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

	id.SetLogger(logger)
	pull.SetLogger(logger)
	push.SetLogger(logger)

	dataconnector.Inject(dataconnectorStorage(), dataPingerFactory())
	relation.Inject(dataconnectorStorage(), relationStorage(), relationExtractorFactory())
	table.Inject(dataconnectorStorage(), tableStorage(), tableExtractorFactory())
	id.Inject(idStorage(), relationStorage(), idExporter(), idJSONStorage(*os.Stdout))
	pull.Inject(dataconnectorStorage(), relationStorage(), tableStorage(), idStorage(), pullDataSourceFactory(), pullRowExporterFactory(), traceListner(os.Stderr))
	push.Inject(dataconnectorStorage(), relationStorage(), tableStorage(), idStorage(), pushDataDestinationFactory(), pushRowIteratorFactory())
}
