package push

import (
	"reflect"
	"testing"

	"makeit.imfr.cgi.com/lino/pkg/dataconnector"
	"makeit.imfr.cgi.com/lino/pkg/id"
	"makeit.imfr.cgi.com/lino/pkg/push"
	"makeit.imfr.cgi.com/lino/pkg/relation"
	"makeit.imfr.cgi.com/lino/pkg/table"
)

func Test_getDataDestination(t *testing.T) {
	readOnlyDC := dataconnector.DataConnector{Name: "connector-ro", URL: "postgres://localhost/test", ReadOnly: true}
	dcStorage := dataconnector.MockStorage{}
	dcStorage.On("List").Return([]dataconnector.DataConnector{readOnlyDC}, nil)

	Inject(
		&dcStorage,
		&relation.MockStorage{},
		&table.MockStorage{},
		&id.MockStorage{},
		map[string]push.DataDestinationFactory{},
		&push.MockRowIterator{},
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
		struct {
			name  string
			args  args
			want  push.DataDestination
			want1 *push.Error
		}{
			name:  "readonly protect",
			args:  args{dataconnectorName: "connector-ro"},
			want:  nil,
			want1: &push.Error{Description: "'connector-ro' is a read only dataconnector"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getDataDestination(tt.args.dataconnectorName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDataDestination() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getDataDestination() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
