package restmime

import (
	"bytes"
	"io"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/protobuf/proto"
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
			name: "json_withArguments",
			args: args{
				dest: &struct {
					Greeting string `json:"greeting"`
					Name     string `json:"name"`
				}{},
				contentType: "application/json; charset=utf-8",
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
		{
			name: "url_values",
			args: args{
				dest:        &url.Values{},
				contentType: "application/x-www-form-urlencoded",
				reader:      bytes.NewBufferString(`greeting=hi&name=Tom`),
			},
			want:    &url.Values{"greeting": []string{"hi"}, "name": []string{"Tom"}},
			wantErr: false,
		},
		{
			name: "protobuf",
			args: args{
				dest:        &pb.HelloRequest{},
				contentType: "application/x-protobuf",
				reader: func() io.Reader {
					data, _ := proto.Marshal(&pb.HelloRequest{
						Name: "Hi",
					})
					return bytes.NewBuffer(data)
				}(),
			},
			want:    &pb.HelloRequest{Name: "Hi"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal(tt.args.dest, tt.args.contentType, tt.args.reader)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				if assert.Nil(t, err, err) {
					wantMessage, _ := tt.want.(proto.Message)
					haveMessage, _ := tt.args.dest.(proto.Message)
					if wantMessage != nil && haveMessage != nil {
						assert.True(t, proto.Equal(wantMessage, haveMessage))
					} else {
						assert.Equal(t, tt.want, tt.args.dest)
					}
				}
			}
		})
	}
}
