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
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"makeit.imfr.cgi.com/lino/pkg/pull"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	var (
		datasource     pull.DataSource
		err            *pull.Error
		datasourceName string
		ok             bool
		filter         map[string]string
		limit          uint
	)

	pathParams := mux.Vars(r)

	query := r.URL.Query()

	filter = map[string]string{}

	if query.Get("filter") != "" {
		for _, f := range strings.Split(query.Get("filter"), ",") {
			kv := strings.SplitN(f, ":", 2)
			if len(kv) != 2 {
				logger.Error("can't parse filter\n")
				w.WriteHeader(http.StatusBadRequest)
				_, ew := w.Write([]byte("{\"error\": \"param filter must be a string map (key1:value1,key2:value2)\"}\n"))
				if ew != nil {
					logger.Error("Write failed\n")
					return
				}
				return
			}
			filter[kv[0]] = kv[1]
		}
	}

	if query.Get("limit") != "" {
		limit64, elimit := strconv.ParseUint(query.Get("limit"), 10, 64)
		if elimit != nil {
			logger.Error("can't parse limie\n")
			w.WriteHeader(http.StatusBadRequest)
			_, ew := w.Write([]byte("{\"error\" : \"param limit must be an positive integer\"}\n"))
			if ew != nil {
				logger.Error("Write failed\n")
				return
			}
			return
		}
		limit = uint(limit64)
	}

	w.Header().Set("Content-Type", "application/json")

	if datasourceName, ok = pathParams["dataSource"]; !ok {
		logger.Error("param datasource is required\n")
		w.WriteHeader(http.StatusBadRequest)
		_, ew := w.Write([]byte("{\"error\": \"param datasource is required\"}"))
		if ew != nil {
			logger.Error("Write failed\n")
			return
		}
		return
	}

	datasource, err = getDataSource(datasourceName, w)
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusNotFound)
		_, ew := w.Write([]byte("{\"error\": \"" + err.Description + "\"}"))
		if ew != nil {
			logger.Error("Write failed\n")
			return
		}
		return
	}

	plan, e2 := getPullerPlan(filter, limit, idStorageFactory(query.Get("table")))
	if e2 != nil {
		logger.Error(e2.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, ew := w.Write([]byte("{\"error\": \"" + e2.Description + "}"))
		if ew != nil {
			logger.Error("Write failed\n")
			return
		}
		return
	}

	pullExporter := pullExporterFactory(w)

	e3 := pull.Pull(plan, pull.NewOneEmptyRowReader(), datasource, pullExporter, pull.NoTraceListener{})
	if e3 != nil {
		logger.Error(e3.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, ew := w.Write([]byte(e3.Description))
		if ew != nil {
			logger.Error("Write failed\n")
			return
		}
		return
	}
}
