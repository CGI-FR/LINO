package pull

import (
	"strings"
	"testing"

	"makeit.imfr.cgi.com/lino/pkg/pull"
)

func TestJSONRowReader_Next(t *testing.T) {
	tests := []struct {
		name   string
		stream string
		want   []pull.Row
	}{
		{
			"simple",
			"{\"name\": \"test\"}\n",
			[]pull.Row{{"name": "test"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jrr := NewJSONRowReader(strings.NewReader(tt.stream))

			for _, row := range tt.want {
				if got := jrr.Next(); got != true {
					t.Errorf("JSONRowReader.Next() = %v, want %v", got, true)
				}
				for k, v := range jrr.Value() {
					if row[k] != v {
						t.Errorf("JSONRowReader.Value() = %v, want %v", v, row[k])
					}
				}
			}
		})
	}
}
