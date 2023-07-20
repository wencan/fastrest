package restvalues

import (
	"net/url"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "url_values",
			args: args{
				v: url.Values{"greeting": []string{"hi"}, "name": []string{"Tom"}},
			},
			want: "greeting=hi&name=Tom",
		},
		{
			name: "schema",
			args: args{
				v: struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{
					Greeting: "hi",
					Name:     "Tom",
				},
			},
			want: "greeting=hi&name=Tom",
		},
		{
			name: "schema_ptr",
			args: args{
				v: &struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{
					Greeting: "hi",
					Name:     "Tom",
				},
			},
			want: "greeting=hi&name=Tom",
		},
		{
			name: "time_field",
			args: args{
				v: &struct {
					StartTime time.Time `schema:"start_time"`
				}{
					StartTime: time.Date(2023, 7, 20, 10, 19, 51, 0, time.FixedZone("+08:00", 3600*8)),
				},
			},
			want: "start_time=2023-07-20T10%3A19%3A51%2B08%3A00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encode(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}
