package urlbuilder

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/xo/dburl"
	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
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
	return u
}
