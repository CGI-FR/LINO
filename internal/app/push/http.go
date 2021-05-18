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

package push

import (
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/cgi-fr/lino/pkg/push"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	Handler(w, r, push.Delete)
}

func InsertHandler(w http.ResponseWriter, r *http.Request) {
	Handler(w, r, push.Insert)
}

func TruncatHandler(w http.ResponseWriter, r *http.Request) {
	Handler(w, r, push.Truncate)
}

func Handler(w http.ResponseWriter, r *http.Request, mode push.Mode) {
	pathParams := mux.Vars(r)
	query := r.URL.Query()

	var (
		dcDestination      string
		ok                 bool
		commitSize         = uint(100)
		disableConstraints bool
	)

	if dcDestination, ok = pathParams["dataDestination"]; !ok {
		log.Error().Msg("param dataDestination is required\n")
		w.WriteHeader(http.StatusBadRequest)
		_, ew := w.Write([]byte("{\"error\": \"param dataDestination is required\"}"))
		if ew != nil {
			log.Error().Msg("Write failed\n")
			return
		}
		return
	}

	datadestination, err := getDataDestination(dcDestination)
	if err != nil {
		log.Error().Msg(err.Error())
		w.WriteHeader(http.StatusNotFound)
		_, ew := w.Write([]byte("{\"error\": \"" + html.EscapeString(err.Description) + "\"}"))
		if ew != nil {
			log.Error().Msg("Write failed\n")
			return
		}
		return
	}

	plan, e2 := getPlan(idStorageFactory(query.Get("table")))
	if e2 != nil {
		log.Error().Msg(e2.Error())
		w.WriteHeader(http.StatusNotFound)
		_, ew := w.Write([]byte("{\"error\": \"" + e2.Description + "\"}"))
		if ew != nil {
			log.Error().Msg("Write failed\n")
			return
		}
		return
	}

	if query.Get("commitsize") != "" {
		commitsize64, ecommitsize := strconv.ParseUint(query.Get("commitsize"), 10, 64)
		if ecommitsize != nil {
			log.Error().Msg("can't parse commitsize\n")
			w.WriteHeader(http.StatusBadRequest)
			_, ew := w.Write([]byte("{\"error\" : \"param commitsize must be an positive integer\"}\n"))
			if ew != nil {
				log.Error().Msg("Write failed\n")
				return
			}
			return
		}
		commitSize = uint(commitsize64)
	}

	if query.Get("disable-constraints") != "" {
		var edisableConstraints error
		disableConstraints, edisableConstraints = strconv.ParseBool(query.Get("disable-constraints"))
		if edisableConstraints != nil {
			log.Error().Msg("can't parse disable-constraints\n")
			w.WriteHeader(http.StatusBadRequest)
			_, ew := w.Write([]byte("{\"error\" : \"param disable-constraints must be a boolean\"}\n"))
			if ew != nil {
				log.Error().Msg("Write failed\n")
				return
			}
			return
		}
	}

	log.Debug().Msg(fmt.Sprintf("call Push with mode %s", mode))

	e3 := push.Push(rowIteratorFactory(r.Body), datadestination, plan, mode, commitSize, disableConstraints, push.NoErrorCaptureRowWriter{})
	if e3 != nil {
		log.Error().Msg(e3.Error())
		w.WriteHeader(http.StatusNotFound)
		_, ew := w.Write([]byte("{\"error\": \"" + e3.Description + "\"}"))
		if ew != nil {
			log.Error().Msg("Write failed\n")
			return
		}
		return
	}
	_, ew := w.Write([]byte("{\"error\": \"\"}"))
	if ew != nil {
		log.Error().Msg("Write failed\n")
		return
	}
}
