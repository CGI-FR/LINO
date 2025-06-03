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

	"github.com/cgi-fr/lino/internal/app/pull"
	"github.com/cgi-fr/lino/internal/app/push"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
)

// NewCommand implements the cli http command
func NewCommand(fullName string, err *os.File, out *os.File, in *os.File) *cobra.Command {
	var (
		port              uint
		ingressDescriptor string
		enableCORS        bool
		corsOrigins       []string
		corsMethods       []string
		corsHeaders       []string
	)

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
				HandlerFunc(pull.HandlerFactory(ingressDescriptor))

			api.Path("/data/{dataDestination}").
				Queries("mode", "delete").
				HandlerFunc(push.DeleteHandlerFactory(ingressDescriptor))

			api.Path("/data/{dataDestination}").
				Queries("mode", "insert").
				HandlerFunc(push.InsertHandlerFactory(ingressDescriptor))

			api.Path("/data/{dataDestination}").
				Queries("mode", "truncate").
				HandlerFunc(push.TruncatHandlerFactory(ingressDescriptor))

			api.Path("/data/{dataDestination}").
				Methods(http.MethodPost).
				HandlerFunc(push.TruncatHandlerFactory(ingressDescriptor))

			var handler http.Handler = r

			if enableCORS {
				c := cors.New(cors.Options{
					AllowedOrigins: corsOrigins,
					AllowedMethods: corsMethods,
					AllowedHeaders: corsHeaders,
				})
				handler = c.Handler(r)
			}

			http.Handle("/", handler)
			bind := fmt.Sprintf(":%d", port)
			e1 := http.ListenAndServe(bind, nil) //nolint:gosec

			if err != nil {
				fmt.Fprintln(err, e1.Error())
				os.Exit(1)
			}
		},
	}
	cmd.Flags().UintVarP(&port, "port", "p", 8000, "HTTP Port to bind")
	cmd.Flags().StringVarP(&ingressDescriptor, "ingress-descriptor", "i", "ingress-descriptor.yaml", "Ingress descriptor filename")

	// CORS flags
	cmd.Flags().BoolVar(&enableCORS, "enable-cors", false, "Enable CORS support")
	cmd.Flags().StringSliceVar(&corsOrigins, "cors-origins", []string{"*"}, "Allowed CORS origins (e.g. http://localhost:3000)")
	cmd.Flags().StringSliceVar(&corsMethods, "cors-methods", []string{"GET", "POST", "OPTIONS", "DELETE"}, "Allowed CORS methods")
	cmd.Flags().StringSliceVar(&corsHeaders, "cors-headers", []string{"Content-Type", "Authorization"}, "Allowed CORS headers")

	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetIn(in)
	return cmd
}
