package httpserver

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wencan/fastrest/restcodecs/restmime"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/protobuf/proto"
)

func TestWriteResponse(t *testing.T) {
	type args struct {
		accept   string
		response interface{}
	}
	type want struct {
		header       http.Header
		responseBody []byte
	}
	tests := []struct {
		name      string
		args      args
		want      want
		wantError bool
	}{
		{
			name: "default_json",
			args: args{
				response: struct {
					Echo string `json:"echo"`
				}{Echo: "test"},
			},
			want: want{
				header:       http.Header{"Content-Type": []string{"application/json"}},
				responseBody: []byte("{\"echo\":\"test\"}\n"),
			},
		},
		{
			name: "accept_json",
			args: args{
				accept: restmime.MimeTypeJson,
				response: struct {
					Echo string `json:"echo"`
				}{Echo: "test"},
			},
			want: want{
				header:       http.Header{"Content-Type": []string{"application/json"}},
				responseBody: []byte("{\"echo\":\"test\"}\n"),
			},
		},
		{
			name: "accept_proto",
			args: args{
				accept: restmime.MimeTypeProtobuf,
				response: &helloworld.HelloRequest{
					Name: "Tom",
				},
			},
			want: want{
				header: http.Header{"Content-Type": []string{"application/x-protobuf"}},
				responseBody: func() []byte {
					data, _ := proto.Marshal(&helloworld.HelloRequest{
						Name: "Tom",
					})
					return data
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()

			err := WriteResponse(context.TODO(), http.StatusOK, tt.args.accept, tt.args.response, response)
			if tt.wantError {
				assert.NotNil(t, err)
			} else {
				if assert.Nil(t, err) {

					gotHeader := make(http.Header)
					for key, values := range response.Header() {
						gotHeader[key] = values
					}
					assert.Equal(t, tt.want.header, gotHeader)

					gotBody, _ := io.ReadAll(response.Body)
					if !bytes.Equal(gotBody, tt.want.responseBody) {
						t.Fatalf("want body: %v, got body: %v", string(tt.want.responseBody), string(gotBody))
					}
				}
			}
		})
	}
}
