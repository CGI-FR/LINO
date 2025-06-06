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

package pull

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/cgi-fr/lino/pkg/pull"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func HandlerFactory(ingressDescriptor string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			datasource     pull.DataSource
			err            error
			datasourceName string
			ok             bool
			distinct       bool
			filter         pull.Row
			limit          uint
			where          string
		)

		pathParams := mux.Vars(r)

		query := r.URL.Query()

		filter = pull.Row{}

		if query.Get("filter") != "" {
			for _, f := range strings.Split(query.Get("filter"), ",") {
				kv := strings.SplitN(f, ":", 2)
				if len(kv) != 2 {
					log.Error().Msg("can't parse filter")
					w.WriteHeader(http.StatusBadRequest)
					_, ew := w.Write([]byte("{\"error\": \"param filter must be a string map (key1:value1,key2:value2)\"}\n"))
					if ew != nil {
						log.Error().Msg("Write failed")
						return
					}
					return
				}
				filter[kv[0]] = kv[1]
			}
		}

		if query.Get("limit") != "" {
			limit64, elimit := strconv.ParseUint(query.Get("limit"), 10, 32)
			if elimit != nil || limit64 > math.MaxInt32 {
				log.Error().Msg("can't parse limit or limit is too large")
				w.WriteHeader(http.StatusBadRequest)
				_, ew := w.Write([]byte("{\"error\" : \"param limit must be a positive integer within the valid range\"}\n"))
				if ew != nil {
					log.Error().Msg("Write failed")
					return
				}
				return
			}
			limit = uint(limit64)
		}

		if query.Get("distinct") != "" {
			var edistinct error

			distinct, edistinct = strconv.ParseBool(query.Get("distinct"))
			if edistinct != nil {
				log.Error().Msg("can't parse distinct")
				w.WriteHeader(http.StatusBadRequest)
				_, ew := w.Write([]byte("{\"error\" : \"param distinct must be a boolean\"}\n"))
				if ew != nil {
					log.Error().Msg("Write failed")
					return
				}
				return
			}
		}

		if query.Get("where") != "" {
			where = query.Get("where")
			// CWE-117 : sanitize user input
			where = strings.ReplaceAll(where, "\n", "")
			where = strings.ReplaceAll(where, "\r", "")
		}

		w.Header().Set("Content-Type", "application/json")

		if datasourceName, ok = pathParams["dataSource"]; !ok {
			log.Error().Msg("param datasource is required")
			w.WriteHeader(http.StatusBadRequest)
			_, ew := w.Write([]byte("{\"error\": \"param datasource is required\"}"))
			if ew != nil {
				log.Error().Err(ew).Msg("Write failed")
				return
			}
			return
		}

		datasource, err = getDataSource(datasourceName, w)
		if err != nil {
			log.Error().Err(err).Msg("")
			w.WriteHeader(http.StatusNotFound)
			_, ew := w.Write([]byte("{\"error\": \"" + err.Error() + "\"}"))
			if ew != nil {
				log.Error().Err(ew).Msg("Write failed")
				return
			}
			return
		}

		plan, start, startSelect, e2 := getPullerPlan(idStorageFactory(query.Get("table"), ingressDescriptor))
		if e2 != nil {
			log.Error().Err(e2).Msg("")
			w.WriteHeader(http.StatusInternalServerError)
			_, ew := w.Write([]byte("{\"error\": \"" + e2.Error() + "}"))
			if ew != nil {
				log.Error().Err(ew).Msg("Write failed")
				return
			}
			return
		}

		pullExporter := pullExporterFactory(w)
		puller := pull.NewPuller(plan, datasource, pullExporter, pull.NoTraceListener{})

		e3 := puller.Pull(start, pull.Filter{Limit: limit, Values: filter, Where: where, Distinct: distinct}, startSelect, nil, nil)
		if e3 != nil {
			log.Error().Err(e3).Msg("")
			w.WriteHeader(http.StatusInternalServerError)
			_, ew := w.Write([]byte(e3.Error()))
			if ew != nil {
				log.Error().Err(ew).Msg("Write failed")
				return
			}
			return
		}
	}
}
