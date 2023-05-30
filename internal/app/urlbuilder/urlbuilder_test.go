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
			`oracle://user:pwd@dbhost:1521/orclpdb1/`,
		},
		{
			"oracle-raw",
			"oracle-raw://user:pwd@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=dbhost.example.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=orclpdb1)))",
			`oracle://user:pwd@:0/?connStr=%28DESCRIPTION%3D%28ADDRESS%3D%28PROTOCOL%3DTCP%29%28HOST%3Ddbhost.example.com%29%28PORT%3D1521%29%29%28CONNECT_DATA%3D%28SERVICE_NAME%3Dorclpdb1%29%29%29`,
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
