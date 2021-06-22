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

package urlbuilder

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/cgi-fr/lino/internal/app/localstorage"
	"github.com/cgi-fr/lino/pkg/dataconnector"
	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/xo/dburl"
)

// GenOracleRaw generates a Go Driver for Oracle (godror) DSN from the passed URL.
func GenOracleRaw(u *dburl.URL) (string, error) {
	// example of DSN generated :
	//   `user="login" password="password" connectString="host:port/service_name" sysdba=true`

	connectionString := strings.TrimSuffix(u.Host+u.Path, "/")

	dsn := fmt.Sprintf("connectString=\"%s\"", connectionString)

	for value := range u.Query() {
		dsn += fmt.Sprintf(" %s=\"%s\"", value, u.Query().Get(value))
	}

	// build user/pass
	if u.User != nil {
		if un := u.User.Username(); len(un) > 0 {
			dsn += fmt.Sprintf(" user=\"%s\"", un)
			if up, ok := u.User.Password(); ok {
				dsn += fmt.Sprintf(" password=\"%s\"", up)
			}
		}
	}

	return dsn, nil
}

func init() {
	oracleScheme := dburl.Scheme{
		Driver:    "godror-raw",
		Override:  "godror",
		Generator: GenOracleRaw,
		Opaque:    false,
		Aliases:   []string{"oracle-raw"},
		Proto:     dburl.ProtoAny,
	}
	dburl.Register(oracleScheme)

	db2Scheme := dburl.Scheme{
		Driver: "db2",
		Generator: func(u *dburl.URL) (string, error) {
			password, _ := u.User.Password()
			database := strings.TrimPrefix(u.Path, "/")
			result := fmt.Sprintf("HOSTNAME=%s;DATABASE=%s;PORT=%s;UID=%s;PWD=%s", u.Hostname(), database, u.Port(), u.User.Username(), password)
			return result, nil
		},
		Proto:    dburl.ProtoAny,
		Opaque:   false,
		Aliases:  []string{"go_ibm_db"},
		Override: "go_ibm_db",
	}
	dburl.Register(db2Scheme)
}

func BuildURL(dc *dataconnector.DataConnector, out io.Writer) *dburl.URL {
	u, e2 := dburl.Parse(dc.URL)
	if e2 != nil {
		fmt.Fprintln(out, e2.Error())
		os.Exit(3)
	}
	// get user from env
	if dc.User.ValueFromEnv != "" {
		userFromEnv := os.Getenv(dc.User.ValueFromEnv)
		if userFromEnv == "" {
			if out != nil {
				fmt.Fprintf(out, "warn: missing environment variable %s", dc.User.ValueFromEnv)
				fmt.Fprintln(out)
			}
		} else {
			u.User = url.User(userFromEnv)
		}
	} else if dc.User.Value != "" {
		// set user from dc
		u.User = url.User(dc.User.Value)
	}
	// get password from env
	if dc.Password.ValueFromEnv != "" {
		passwordFromEnv := os.Getenv(dc.Password.ValueFromEnv)
		if passwordFromEnv == "" {
			if out != nil {
				fmt.Fprintf(out, "warn: missing environment variable %s", dc.Password.ValueFromEnv)
				fmt.Fprintln(out)
			}
		} else {
			u.User = url.UserPassword(u.User.Username(), passwordFromEnv)
		}
	}
	// if credentials still missing, check default store
	username := u.User.Username()
	_, passwordIsSet := u.User.Password()
	if username == "" || !passwordIsSet {
		store := defaultCredentialsStore()
		creds, err := client.Get(store, u.String())
		if err != nil {
			// failed to use credential store backend, fallback to local storage
			creds, err = localstorage.Read(u.String())
		}
		if err == nil {
			u.User = url.UserPassword(creds.Username, creds.Secret)
		}
	}
	return u
}

func StorePassword(u *dburl.URL, password string, out io.Writer) error {
	store := defaultCredentialsStore()
	creds := &credentials.Credentials{ServerURL: u.URL.String(), Username: u.URL.User.Username(), Secret: password}
	err := client.Store(store, creds)
	if err != nil {
		// failed to use credential store backend
		if credentials.IsCredentialsMissingServerURL(err) || credentials.IsCredentialsMissingUsername(err) || credentials.IsErrCredentialsNotFound(err) {
			return err
		}
		// fall back to local storage
		fmt.Fprintf(out, "warn: password will be stored unencrypted in %s, configure a credential helper to remove this warning. See https://github.com/docker/docker-credential-helpers", localstorage.GetFileLocation())
		fmt.Fprintln(out)
		return localstorage.Store(creds)
	}
	return nil
}
