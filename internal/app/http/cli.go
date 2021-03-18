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

// NewCommand implements the cli http command
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
				Queries("mode", "delete").
				HandlerFunc(push.DeleteHandler)

			api.Path("/data/{dataDestination}").
				Queries("mode", "insert").
				HandlerFunc(push.InsertHandler)

			api.Path("/data/{dataDestination}").
				Queries("mode", "truncate").
				HandlerFunc(push.TruncatHandler)

			api.Path("/data/{dataDestination}").
				Methods(http.MethodPost).
				HandlerFunc(push.TruncatHandler)

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
