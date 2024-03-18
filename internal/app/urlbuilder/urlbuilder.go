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
	go_ora "github.com/sijms/go-ora/v2"
	"github.com/xo/dburl"
)

// GenOracleRaw generates a Go Driver for Oracle (godror) DSN from the passed URL.
func GenOracleRaw(u *dburl.URL) (string, string, error) {
	// example of DSN generated :
	//   `oracle://proxy_user:proxy_password@host:port/service?proxy client name=schema_owner`

	dsn := "oracle://"

	if strings.HasPrefix(u.Hostname(), "(") {
		// build user/pass
		return genJDBCOracle(u)
	}

	// build user/pass
	if u.User != nil {
		if un := u.User.Username(); len(un) > 0 {
			dsn += un
			if up, ok := u.User.Password(); ok {
				dsn += fmt.Sprintf(":%s", up)
			}
		}
		dsn += "@"
	}

	dsn += fmt.Sprintf("%s%s?", u.Host, u.Path)

	for value := range u.Query() {
		dsn += fmt.Sprintf("%s=\"%s\"", value, u.Query().Get(value))
	}

	return dsn, "", nil
}

func genJDBCOracle(u *dburl.URL) (string, string, error) {
	if u.User == nil {
		return go_ora.BuildJDBC("", "", u.Host, map[string]string{}), "", nil
	}

	if un := u.User.Username(); len(un) > 0 {
		if up, ok := u.User.Password(); ok {
			return go_ora.BuildJDBC(u.User.Username(), up, u.Host, map[string]string{}), "", nil
		} else {
			return go_ora.BuildJDBC(u.User.Username(), "", u.Host, map[string]string{}), "", nil
		}
	}

	return go_ora.BuildJDBC("", "", u.Host, map[string]string{}), "", nil
}

func genMySQL(u *dburl.URL) (string, string, error) {
	dsn, other, err := dburl.GenMysql(u)
	if err != nil {
		return dsn, other, err
	}

	return strings.ReplaceAll(dsn, "/", "?"), other, err
}

func init() {
	oracleScheme := dburl.Scheme{
		Driver:    "godror-raw",
		Generator: GenOracleRaw,
		Transport: dburl.TransportAny,
		Opaque:    false,
		Aliases:   []string{"oracle-raw"},
		Override:  "oracle",
	}
	dburl.Register(oracleScheme)

	db2Scheme := dburl.Scheme{
		Driver: "db2",
		Generator: func(u *dburl.URL) (string, string, error) {
			password, _ := u.User.Password()
			database := strings.TrimPrefix(u.Path, "/")
			result := fmt.Sprintf("HOSTNAME=%s;DATABASE=%s;PORT=%s;UID=%s;PWD=%s", u.Hostname(), database, u.Port(), u.User.Username(), password)
			return result, "", nil
		},
		Transport: dburl.TransportAny,
		Opaque:    false,
		Aliases:   []string{"go_ibm_db"},
		Override:  "go_ibm_db",
	}
	dburl.Register(db2Scheme)

	httpScheme := dburl.Scheme{
		Driver:    "http",
		Generator: dburl.GenFromURL("http://localhost:8080"),
		Transport: dburl.TransportAny,
		Opaque:    false,
		Aliases:   []string{"https"},
		Override:  "",
	}
	dburl.Register(httpScheme)

	wsScheme := dburl.Scheme{
		Driver:    "ws",
		Generator: dburl.GenFromURL("ws://localhost:8080"),
		Transport: dburl.TransportAny,
		Opaque:    false,
		Aliases:   []string{"wss"},
		Override:  "",
	}
	dburl.Register(wsScheme)

	mySQLScheme := dburl.Scheme{
		Driver:    "mysql-test",
		Generator: genMySQL,
		Transport: dburl.TransportAny,
		Opaque:    false,
		Aliases:   []string{},
		Override:  "",
	}
	dburl.Register(mySQLScheme)
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
