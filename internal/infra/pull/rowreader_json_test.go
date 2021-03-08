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
