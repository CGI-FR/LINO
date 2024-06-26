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
	"bytes"
	"fmt"
	netHttp "net/http"
	"os"
	"runtime"
	"strings"
	"text/template"

	over "github.com/adrienaury/zeromdc"
	"github.com/cgi-fr/lino/internal/app/analyse"
	"github.com/cgi-fr/lino/internal/app/dataconnector"
	"github.com/cgi-fr/lino/internal/app/http"
	"github.com/cgi-fr/lino/internal/app/id"
	"github.com/cgi-fr/lino/internal/app/pull"
	"github.com/cgi-fr/lino/internal/app/push"
	"github.com/cgi-fr/lino/internal/app/query"
	"github.com/cgi-fr/lino/internal/app/relation"
	"github.com/cgi-fr/lino/internal/app/sequence"
	"github.com/cgi-fr/lino/internal/app/table"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Provisioned by ldflags
// nolint: gochecknoglobals
var (
	version   string
	commit    string
	buildDate string
	builtBy   string

	// global flags
	loglevel            string
	jsonlog             bool
	debug               bool
	colormode           string
	statsDestination    string
	statsTemplate       string
	statsDestinationEnv = os.Getenv("LINO_STATS_URL")
	statsTemplateEnv    = os.Getenv("LINO_STATS_TEMPLATE")
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lino [action]",
	Short: "Command line tools for managing tests data",
	Long: `Lino is a simple ETL (Extract Transform Load) tools to manage tests datas. The lino command line tool pull test data from a relational database to create a smallest production-like database.

Environment Variables:
  LINO_STATS_URL      The URL where statistics will be sent
  LINO_STATS_TEMPLATE The template string to format statistics`,
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Info().
			Str("verbosity", loglevel).
			Bool("log-json", jsonlog).
			Bool("debug", debug).
			Str("color", colormode).
			Msg("Start LINO")
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		stats, ok := over.MDC().Get("stats")
		if ok {
			statsByte := stats.([]byte)
			log.Info().RawJSON("stats", statsByte).Int("return", 0).Msg("End LINO")

			statsToWrite := statsByte
			if statsTemplate != "" {
				tmpl, err := template.New("statsTemplate").Parse(statsTemplate)
				if err != nil {
					log.Error().Err(err).Msg(("Error parsing statistics template"))
					os.Exit(1)
				}
				var output bytes.Buffer
				err = tmpl.ExecuteTemplate(&output, "statsTemplate", Stats{Stats: string(statsByte)})
				if err != nil {
					log.Error().Err(err).Msg("Error adding stats to template")
					os.Exit(1)
				}
				statsToWrite = output.Bytes()
			}

			if statsDestination != "" {
				if strings.HasPrefix(statsDestination, "http") {
					sendMetrics(statsDestination, statsToWrite)
				} else {
					writeMetricsToFile(statsDestination, statsToWrite)
				}
			}
		} else {
			log.Info().Int("return", 0).Msg("End LINO")
		}
	},
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
	rootCmd.PersistentFlags().StringVarP(&loglevel, "verbosity", "v", "error", "set level of log verbosity : none (0), error (1), warn (2), info (3), debug (4), trace (5)")
	rootCmd.PersistentFlags().BoolVar(&jsonlog, "log-json", false, "output logs in JSON format")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "add debug information to logs (very slow)")
	rootCmd.PersistentFlags().StringVar(&colormode, "color", "auto", "use colors in log outputs : yes, no or auto")
	rootCmd.PersistentFlags().StringVar(&statsDestination, "stats", statsDestinationEnv, "file to output statistics to")
	rootCmd.PersistentFlags().StringVar(&statsTemplate, "statsTemplate", statsTemplateEnv, "template string to format stats (to include them you have to specify them as `{{ .Stats }}` like `{\"software\":\"LINO\",\"stats\":{{ .Stats }}}`)")
	rootCmd.AddCommand(dataconnector.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(table.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(sequence.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(relation.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(id.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(pull.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(push.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(http.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(analyse.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
	rootCmd.AddCommand(query.NewCommand("lino", os.Stderr, os.Stdout, os.Stdin))
}

func initConfig() {
	color := false
	switch strings.ToLower(colormode) {
	case "auto":
		if isatty.IsTerminal(os.Stdout.Fd()) && runtime.GOOS != "windows" {
			color = true
		}
	case "yes", "true", "1", "on", "enable":
		color = true
	}

	var logger zerolog.Logger

	if jsonlog {
		logger = zerolog.New(os.Stderr)
	} else {
		logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: !color})
	}
	if debug {
		logger = logger.With().Caller().Logger()
	}

	over.New(logger)

	switch loglevel {
	case "trace", "5":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		log.Info().Msg("Logger level set to trace")
	case "debug", "4":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Info().Msg("Logger level set to debug")
	case "info", "3":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Info().Msg("Logger level set to info")
	case "warn", "2":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error", "1":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	analyse.Inject(tableStorage(), dataconnectorStorage(), analyseDataSourceFactory())
	dataconnector.Inject(dataconnectorStorage(), dataPingerFactory())
	relation.Inject(dataconnectorStorage(), relationStorage(), relationExtractorFactory())
	table.Inject(dataconnectorStorage(), tableStorage(), tableExtractorFactory())
	sequence.Inject(dataconnectorStorage(), tableStorage(), sequenceStorage(), sequenceUpdatorFactory())
	id.Inject(idStorageFile, relationStorage(), idExporter(), idJSONStorage(*os.Stdout))
	pull.Inject(dataconnectorStorage(), relationStorage(), tableStorage(), idStorageFactory(), pullDataSourceFactory(), pullRowExporterFactory(), pullRowReaderFactory(), pullKeyStoreFactory(), traceListner(os.Stderr))
	push.Inject(dataconnectorStorage(), relationStorage(), tableStorage(), idStorageFactory(), pushDataDestinationFactory(), pushRowIteratorFactory(), pushRowExporterFactory(), pushTranslator(), pushObserver())
	query.Inject(dataconnectorStorage(), queryDataSourceFactory())
}

func writeMetricsToFile(statsFile string, statsByte []byte) {
	file, err := os.Create(statsFile)
	if err != nil {
		log.Error().Err(err).Msg("Error generating statistics dump file")
		os.Exit(1)
	}
	defer file.Close()

	_, err = file.Write(statsByte)
	if err != nil {
		log.Error().Err(err).Msg("Error writing statistics to dump file")
	}
	log.Info().Msgf("Statistics exported to file %s", file.Name())
}

func sendMetrics(statsDestination string, statsByte []byte) {
	requestBody := bytes.NewBuffer(statsByte)
	// nolint: gosec
	_, err := netHttp.Post(statsDestination, "application/json", requestBody)
	if err != nil {
		log.Error().Err(err).Msgf("An error occurred trying to send metrics to %s", statsDestination)
	}
	log.Info().Msgf("Statistics sent to %s", statsDestination)
}

type Stats struct {
	Stats interface{} `json:"stats"`
}
