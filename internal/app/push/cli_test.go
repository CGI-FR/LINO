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

package push

import (
	"io"
	"reflect"
	"testing"

	"github.com/cgi-fr/lino/pkg/dataconnector"
	"github.com/cgi-fr/lino/pkg/id"
	"github.com/cgi-fr/lino/pkg/push"
	"github.com/cgi-fr/lino/pkg/relation"
	"github.com/cgi-fr/lino/pkg/table"
)

func Test_getDataDestination(t *testing.T) {
	readOnlyDC := dataconnector.DataConnector{Name: "connector-ro", URL: "postgres://localhost/test", ReadOnly: true}
	dcStorage := dataconnector.MockStorage{}
	dcStorage.On("List").Return([]dataconnector.DataConnector{readOnlyDC}, nil)

	Inject(
		&dcStorage,
		&relation.MockStorage{},
		&table.MockStorage{},
		func(string, string) id.Storage { return &id.MockStorage{} },
		map[string]push.DataDestinationFactory{},
		func(io.ReadCloser) push.RowIterator { return &push.MockRowIterator{} },
		func(io.Writer) push.RowWriter { return &push.MockRowWriter{} },
		push.NewMockTranslator(),
		maxLifeTimeOption,
		maxOpenConnsOption,
		maxIdleConnsOption,
	)

	type args struct {
		dataconnectorName string
	}
	tests := []struct {
		name  string
		args  args
		want  push.DataDestination
		want1 *push.Error
	}{
		{
			name:  "readonly protect",
			args:  args{dataconnectorName: "connector-ro"},
			want:  nil,
			want1: &push.Error{Description: "'connector-ro' is a read only dataconnector"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getDataDestination(tt.args.dataconnectorName, -1, -1, -1)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDataDestination() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getDataDestination() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
