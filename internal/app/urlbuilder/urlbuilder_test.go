package urlbuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xo/dburl"
)

func TestBuildOracleURL(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			"oracle-raw",
			"oracle-raw://user:pwd@dbhost:1521/orclpdb1/?connect_timeout=2",
			`connectString="dbhost:1521/orclpdb1" connect_timeout="2" user="user" password="pwd"`,
		},
		{
			"oracle-raw",
			"oracle-raw://user:pwd@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=dbhost.example.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=orclpdb1)))",
			`connectString="(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=dbhost.example.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=orclpdb1)))" user="user" password="pwd"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := dburl.Parse(tt.args)
			if err != nil {
				t.Errorf("parse return error : %v", err)
			}

			assert.Equal(t, tt.want, url.DSN)
		})
	}
}
