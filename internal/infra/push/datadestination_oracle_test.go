package push

import (
	"reflect"
	"testing"

	"github.com/cgi-fr/lino/pkg/push"
	_ "github.com/godror/godror"
)

func TestOracleDialect_ConvertValue(t *testing.T) {
	tests := []struct {
		name string
		from push.Value
		want push.Value
	}{
		{"nil", nil, nil},
		{"string", "Grenoble", "Grenoble"},
		{"integer", 1, 1},
		// convert JSON date to Oracle date format as a workaround for godror
		// https://github.com/godror/godror#timestamp
		{"string json date", "2012-04-23T18:25:43.511Z", "23-Apr-12 6:25:43.511000 PM"},
		{"string date other format", "2010 04 23 18:25:43.511Z", "2010 04 23 18:25:43.511Z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := OracleDialect{}
			if got := d.ConvertValue(tt.from); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OracleDialect.ConvertValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
