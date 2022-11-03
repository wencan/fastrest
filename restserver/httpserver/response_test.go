package httpserver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wencan/fastrest/restcodecs/restmime"
	"github.com/wencan/fastrest/resterror"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/protobuf/proto"
)

func TestWriteResponse(t *testing.T) {
	type args struct {
		r        *http.Request
		response interface{}
		err      error
	}
	type want struct {
		statusCode   int
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
			name: "default_json", // no Accept
			args: args{
				r: func() *http.Request {
					r, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/test", nil)
					return r
				}(),
				response: struct {
					Echo string `json:"echo"`
				}{Echo: "test"},
			},
			want: want{
				statusCode:   http.StatusOK,
				header:       http.Header{"Content-Type": []string{"application/json"}},
				responseBody: []byte("{\"echo\":\"test\"}\n"),
			},
		},
		{
			name: "accept_json",
			args: args{
				r: func() *http.Request {
					r, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/test", nil)
					r.Header.Set("Accept", restmime.MimeTypeJson)
					return r
				}(),
				response: struct {
					Echo string `json:"echo"`
				}{Echo: "test"},
			},
			want: want{
				statusCode:   http.StatusOK,
				header:       http.Header{"Content-Type": []string{"application/json"}},
				responseBody: []byte("{\"echo\":\"test\"}\n"),
			},
		},
		{
			name: "accept_proto",
			args: args{
				r: func() *http.Request {
					r, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/test", nil)
					r.Header.Set("Accept", restmime.MimeTypeProtobuf)
					return r
				}(),
				response: &helloworld.HelloRequest{
					Name: "Tom",
				},
			},
			want: want{
				statusCode: http.StatusOK,
				header:     http.Header{"Content-Type": []string{"application/x-protobuf"}},
				responseBody: func() []byte {
					data, _ := proto.Marshal(&helloworld.HelloRequest{
						Name: "Tom",
					})
					return data
				}(),
			},
		},
		{
			name: "503",
			args: args{
				r: func() *http.Request {
					r, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/test", nil)
					r.Header.Set("Accept", restmime.MimeTypeJson)
					return r
				}(),
				response: nil,
				err:      resterror.ErrorWithStatus(errors.New("test"), resterror.StatusUnavailable),
			},
			want: want{
				statusCode: http.StatusServiceUnavailable,
				header:     http.Header{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()

			err := WriteResponse(context.TODO(), response, tt.args.r, tt.args.response, tt.args.err)
			if tt.wantError {
				assert.NotNil(t, err)
			} else {
				if assert.Nil(t, err) {
					assert.Equal(t, tt.want.statusCode, response.Code)

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
