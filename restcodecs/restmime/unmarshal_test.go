package restmime

import (
	"bytes"
	"io"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	type args struct {
		dest        interface{}
		contentType string
		reader      io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "json",
			args: args{
				dest: &struct {
					Greeting string `json:"greeting"`
					Name     string `json:"name"`
				}{},
				contentType: "application/json",
				reader:      bytes.NewBufferString(`{"greeting":"hi","name":"Tom"}`),
			},
			want: &struct {
				Greeting string `json:"greeting"`
				Name     string `json:"name"`
			}{
				Greeting: "hi",
				Name:     "Tom",
			},
			wantErr: false,
		},
		{
			name: "form",
			args: args{
				dest: &struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{},
				contentType: "application/x-www-form-urlencoded",
				reader:      bytes.NewBufferString(`greeting=hi&name=Tom`),
			},
			want: &struct {
				Greeting string `schema:"greeting"`
				Name     string `schema:"name"`
			}{
				Greeting: "hi",
				Name:     "Tom",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unmarshal(tt.args.dest, tt.args.contentType, tt.args.reader); (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
