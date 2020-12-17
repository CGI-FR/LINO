package push

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"makeit.imfr.cgi.com/lino/pkg/push"
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
		logger.Error("param dataDestination is required\n")
		w.WriteHeader(http.StatusBadRequest)
		_, ew := w.Write([]byte("{\"error\": \"param dataDestination is required\"}"))
		if ew != nil {
			logger.Error("Write failed\n")
			return
		}
		return
	}

	datadestination, err := getDataDestination(dcDestination)
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

	plan, e2 := getPlan(idStorageFactory(query.Get("table")))
	if e2 != nil {
		logger.Error(e2.Error())
		w.WriteHeader(http.StatusNotFound)
		_, ew := w.Write([]byte("{\"error\": \"" + e2.Description + "\"}"))
		if ew != nil {
			logger.Error("Write failed\n")
			return
		}
		return
	}

	if query.Get("commitsize") != "" {
		commitsize64, ecommitsize := strconv.ParseUint(query.Get("commitsize"), 10, 64)
		if ecommitsize != nil {
			logger.Error("can't parse commitsize\n")
			w.WriteHeader(http.StatusBadRequest)
			_, ew := w.Write([]byte("{\"error\" : \"param commitsize must be an positive integer\"}\n"))
			if ew != nil {
				logger.Error("Write failed\n")
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
			logger.Error("can't parse disable-constraints\n")
			w.WriteHeader(http.StatusBadRequest)
			_, ew := w.Write([]byte("{\"error\" : \"param disable-constraints must be a boolean\"}\n"))
			if ew != nil {
				logger.Error("Write failed\n")
				return
			}
			return
		}
	}

	logger.Debug(fmt.Sprintf("call Push with mode %s", mode))

	e3 := push.Push(rowIteratorFactory(r.Body), datadestination, plan, mode, commitSize, disableConstraints, push.NoErrorCaptureRowWriter{})
	if e3 != nil {
		logger.Error(e3.Error())
		w.WriteHeader(http.StatusNotFound)
		_, ew := w.Write([]byte("{\"error\": \"" + e3.Description + "\"}"))
		if ew != nil {
			logger.Error("Write failed\n")
			return
		}
		return
	}
	_, ew := w.Write([]byte("{\"error\": \"\"}"))
	if ew != nil {
		logger.Error("Write failed\n")
		return
	}
}
