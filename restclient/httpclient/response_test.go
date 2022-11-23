package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wencan/fastrest/restcodecs/restmime"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/protobuf/proto"
)

func TestReadResponseBody(t *testing.T) {
	type args struct {
		dest       interface{}
		statusCode int
		header     http.Header
		body       []byte
	}
	type want struct {
		dest interface{}
		err  bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "read_post_json",
			args: args{
				dest: &struct {
					Greeting string `json:"greeting"`
					Name     string `json:"name"`
				}{},
				statusCode: http.StatusOK,
				header: http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
				},
				body: []byte("{\"greeting\":\"hi\",\"name\":\"Tom\"}\n"),
			},
			want: want{
				dest: &struct {
					Greeting string `json:"greeting"`
					Name     string `json:"name"`
				}{
					Greeting: "hi",
					Name:     "Tom",
				},
			},
		},
		{
			name: "read_post_form",
			args: args{
				dest: &struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{},
				statusCode: http.StatusOK,
				header: http.Header{
					"Content-Type": []string{"application/x-www-form-urlencoded"},
				},
				body: []byte(`greeting=hi&name=Tom`),
			},
			want: want{
				dest: &struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{
					Greeting: "hi",
					Name:     "Tom",
				},
			},
		},
		{
			name: "read_post_to_values",
			args: args{
				dest:       &url.Values{},
				statusCode: http.StatusOK,
				header: http.Header{
					"Content-Type": []string{"application/x-www-form-urlencoded"},
				},
				body: []byte(`greeting=hi&name=Tom`),
			},
			want: want{
				dest: &url.Values{"greeting": []string{"hi"}, "name": []string{"Tom"}},
			},
		},
		{
			name: "read_post_protobuf",
			args: args{
				dest:       &pb.HelloRequest{},
				statusCode: http.StatusOK,
				header: http.Header{
					"Content-Type": []string{restmime.MimeTypeProtobuf},
				},
				body: func() []byte {
					data, _ := proto.Marshal(&pb.HelloRequest{
						Name: "Hi",
					})
					return data
				}(),
			},
			want: want{
				dest: &pb.HelloRequest{
					Name: "Hi",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for key, values := range tt.args.header {
					w.Header()[key] = values
				}
				w.WriteHeader(tt.args.statusCode)
				if tt.args.body != nil {
					w.Write(tt.args.body)
				}
			}))
			defer s.Close()

			response, err := s.Client().Get(s.URL + "/test")
			if assert.Nil(t, err) {
				err := ReadResponseBody(context.TODO(), tt.args.dest, response)
				if tt.want.err {
					assert.NotNil(t, err)
				} else {
					if assert.Nil(t, err) {
						haveMessage, _ := tt.args.dest.(proto.Message)
						wantMessage, _ := tt.want.dest.(proto.Message)
						if haveMessage != nil && wantMessage != nil {
							assert.True(t, proto.Equal(haveMessage, wantMessage))
						} else {
							assert.Equal(t, tt.want.dest, tt.args.dest)
						}
					}
				}
			}
		})
	}
}
