package urlbuilder

import (
	"testing"

	"github.com/xo/dburl"
)

// TODO : fix this test
func TestBuildOracleURL(t *testing.T) {
	t.Skip("WARNING : skipping test")
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			"oracle",
			"oracle://user:pwd@dbhost:1521/orclpdb1/?connect_timeout=2",
			"user/pwd@//dbhost:1521/orclpdb1",
		},
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
		{
			"oracle",
			"oracle://user:pwd@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=dbhost.example.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=orclpdb1)))",
			"user/pwd@//(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=dbhost.example.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=orclpdb1)))",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := dburl.Parse(tt.args)
			if err != nil {
				t.Errorf("parse return error : %v", err)
			}

			if url.DSN != tt.want {
				t.Errorf("DSN should be %s not %v", tt.want, url.DSN)
			}
		})
	}
}
