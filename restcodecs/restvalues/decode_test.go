package restvalues

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	type args struct {
		dest interface{}
		str  string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "url_values",
			args: args{
				dest: &url.Values{},
				str:  "greeting=hi&name=Tom",
			},
			want: &url.Values{"greeting": []string{"hi"}, "name": []string{"Tom"}},
		},
		{
			name: "schema",
			args: args{
				dest: &struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{},
				str: "greeting=hi&name=Tom",
			},
			want: &struct {
				Greeting string `schema:"greeting"`
				Name     string `schema:"name"`
			}{
				Greeting: "hi",
				Name:     "Tom",
			},
		},
		{
			name: "time_field",
			args: args{
				dest: &struct {
					StartTime time.Time `schema:"start_time"`
				}{},
				str: "start_time=2023-07-20T10%3A19%3A51Z",
			},
			want: &struct {
				StartTime time.Time `schema:"start_time"`
			}{
				StartTime: time.Date(2023, 7, 20, 10, 19, 51, 0, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Decode(tt.args.dest, tt.args.str)
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Equal(t, tt.want, tt.args.dest)
		})
	}
}
