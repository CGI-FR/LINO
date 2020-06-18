package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"makeit.imfr.cgi.com/lino/internal/app/pull"
	"makeit.imfr.cgi.com/lino/internal/app/push"
)

// NewCommand implements the cli pull command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	var port uint

	cmd := &cobra.Command{
		Use:     "http",
		Short:   "Start HTTP server",
		Long:    "",
		Example: fmt.Sprintf("  %[1]s http --port 8080", fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			r := mux.NewRouter()

			api := r.PathPrefix("/api/v1").Subrouter()

			api.Path("/data/{dataSource}").
				Methods(http.MethodGet).
				HandlerFunc(pull.Handler)

			api.Path("/data/{dataDestination}").
				Methods(http.MethodDelete).
				HandlerFunc(push.DeleteHandler)

			http.Handle("/", r)
			bind := fmt.Sprintf(":%d", port)
			e1 := http.ListenAndServe(bind, nil)

			if err != nil {
				fmt.Fprintln(err, e1.Error())
				os.Exit(1)
			}
		},
	}
	cmd.Flags().UintVarP(&port, "port", "p", 8000, "HTTP Port to bind")
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
