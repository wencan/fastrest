package restvalues

import (
	"net/url"
	"testing"
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
