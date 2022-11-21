package restmime

import (
	"bytes"
	"io"
	"net/url"
	"testing"

	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/protobuf/proto"
)

func TestMarshal(t *testing.T) {
	type args struct {
		v           interface{}
		contentType string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "json",
			args: args{
				v: &struct {
					Greeting string `json:"greeting"`
					Name     string `json:"name"`
				}{
					Greeting: "hi",
					Name:     "Tom",
				},
				contentType: "application/json",
			},
			want:    []byte("{\"greeting\":\"hi\",\"name\":\"Tom\"}\n"),
			wantErr: false,
		},
		{
			name: "form",
			args: args{
				v: &struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{
					Greeting: "hi",
					Name:     "Tom",
				},
				contentType: "application/x-www-form-urlencoded",
			},
			want:    []byte(`greeting=hi&name=Tom`),
			wantErr: false,
		},
		{
			name: "url_values",
			args: args{
				v:           url.Values{"greeting": []string{"hi"}, "name": []string{"Tom"}},
				contentType: "application/x-www-form-urlencoded",
			},
			want:    []byte(`greeting=hi&name=Tom`),
			wantErr: false,
		},
		{
			name: "protobuf",
			args: args{
				v: &pb.HelloRequest{
					Name: "Hi",
				},
				contentType: "application/x-protobuf",
			},
			want: func() []byte {
				data, _ := proto.Marshal(&pb.HelloRequest{
					Name: "Hi",
				})
				return data
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			if err := Marshal(tt.args.v, tt.args.contentType, &buffer); (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			data, err := io.ReadAll(&buffer)
			if err != nil {
				t.Fatal(err)
				return
			}
			if !bytes.Equal(tt.want, data) {
				t.Errorf("Marshal() = %v, want %v", data, tt.want)
			}
		})
	}
}

func TestAcceptableMarshalContentType(t *testing.T) {
	type args struct {
		accept string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "accept_json",
			args: args{
				accept: "application/json",
			},
			want: MimeTypeJson,
		},
		{
			name: "accept_json_with_args",
			args: args{
				accept: "application/json; charset=utf-8",
			},
			want: MimeTypeJson,
		},
		{
			name: "accept_application",
			args: args{
				accept: "application/*",
			},
			want: MimeTypeJson,
		},
		{
			name: "accept_any",
			args: args{
				accept: "*/*",
			},
			want: MimeTypeJson,
		},
		{
			name: "use_first",
			args: args{
				accept: "text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8",
			},
			want: MimeTypeJson,
		},
		{
			name: "use_second",
			args: args{
				accept: "application/xml, application/json;q=0.9, */*;q=0.8",
			},
			want: MimeTypeJson,
		},
		{
			name: "missmatch",
			args: args{
				accept: "text/html, application/xhtml+xml, application/xml;q=0.9",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AcceptableMarshalContentType(tt.args.accept); got != tt.want {
				t.Errorf("AcceptableMarshalContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}
