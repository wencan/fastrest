package httpclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wencan/fastrest/restcodecs/restmime"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/protobuf/proto"
)

func TestNewRequestWithQuery(t *testing.T) {
	type args struct {
		method string
		url    string
		query  interface{}
	}
	type want struct {
		method string
		url    string
		header http.Header
		body   []byte
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "get",
			args: args{
				method: http.MethodGet,
				url:    "/123/index.html",
			},
			want: want{
				method: http.MethodGet,
				url:    "/123/index.html",
				header: http.Header{},
				body:   []byte{},
			},
		},
		{
			name: "get_values",
			args: args{
				method: http.MethodGet,
				url:    "/123/index.html",
				query:  url.Values{"name": []string{"Tom"}, "age": []string{"18"}},
			},
			want: want{
				method: http.MethodGet,
				url:    "/123/index.html?age=18&name=Tom",
				header: http.Header{},
				body:   []byte{},
			},
		},
		{
			name: "get_schema",
			args: args{
				method: http.MethodGet,
				url:    "/123/index.html",
				query: struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{
					Greeting: "hi",
					Name:     "Tom",
				},
			},
			want: want{
				method: http.MethodGet,
				url:    "/123/index.html?greeting=hi&name=Tom",
				header: http.Header{},
				body:   []byte{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod string
			var gotUrl string
			var gotHeader http.Header
			var gotBody []byte

			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotUrl = r.RequestURI
				gotHeader = make(http.Header)
				for key, wantValues := range tt.want.header {
					gotHeader[key] = wantValues
				}
				gotBody, _ = io.ReadAll(r.Body)
				r.Body.Close()
			}))
			defer s.Close()

			r, err := NewRequestWithQuery(context.TODO(), tt.args.method, s.URL+tt.args.url, tt.args.query)
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)

			response, err := s.Client().Do(r)
			if assert.Nil(t, err, err) {
				assert.Equal(t, response.StatusCode, http.StatusOK)

				assert.Equal(t, tt.want.method, gotMethod)
				assert.Equal(t, tt.want.url, gotUrl)
				assert.Equal(t, tt.want.header, gotHeader)
				assert.Equal(t, tt.want.body, gotBody)
			}

		})
	}
}

func TestNewRequestWithBody(t *testing.T) {
	type args struct {
		method      string
		url         string
		contentType string
		bodyObj     interface{}
	}
	type want struct {
		method string
		url    string
		header http.Header
		body   []byte
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "post_json",
			args: args{
				method:      http.MethodPost,
				url:         "/123/abc/json",
				contentType: restmime.MimeTypeJson,
				bodyObj: struct {
					Greeting string `json:"greeting"`
					Name     string `json:"name"`
				}{
					Greeting: "hi",
					Name:     "Tom",
				},
			},
			want: want{
				method: http.MethodPost,
				url:    "/123/abc/json",
				header: http.Header{
					"Content-Type": []string{restmime.MimeTypeJson},
				},
				body: []byte("{\"greeting\":\"hi\",\"name\":\"Tom\"}\n"),
			},
		},
		{
			name: "put_form",
			args: args{
				method:      http.MethodPut,
				url:         "/123/abc/form",
				contentType: restmime.MimeTypeForm,
				bodyObj: struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{
					Greeting: "hi",
					Name:     "Tom",
				},
			},
			want: want{
				method: http.MethodPut,
				url:    "/123/abc/form",
				header: http.Header{
					"Content-Type": []string{restmime.MimeTypeForm},
				},
				body: []byte("greeting=hi&name=Tom"),
			},
		},
		{
			name: "post_values",
			args: args{
				method:      http.MethodPost,
				url:         "/123/abc/values",
				contentType: restmime.MimeTypeForm,
				bodyObj: url.Values{
					"greeting": []string{"hi"},
					"name":     []string{"Tom"},
				},
			},
			want: want{
				method: http.MethodPost,
				url:    "/123/abc/values",
				header: http.Header{
					"Content-Type": []string{restmime.MimeTypeForm},
				},
				body: []byte("greeting=hi&name=Tom"),
			},
		},
		{
			name: "post_protobuf",
			args: args{
				method:      http.MethodPost,
				url:         "/123/proto/buf",
				contentType: restmime.MimeTypeProtobuf,
				bodyObj: &pb.HelloRequest{
					Name: "Hi",
				},
			},
			want: want{
				method: http.MethodPost,
				url:    "/123/proto/buf",
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod string
			var gotUrl string
			var gotHeader http.Header
			var gotBody []byte

			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotUrl = r.RequestURI
				gotHeader = make(http.Header)
				for key, wantValues := range tt.want.header {
					gotHeader[key] = wantValues
				}
				gotBody, _ = io.ReadAll(r.Body)
				r.Body.Close()
			}))
			defer s.Close()

			r, err := NewRequestWithBody(context.TODO(), tt.args.method, s.URL+tt.args.url, tt.args.contentType, tt.args.bodyObj)
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			if !assert.Nil(t, err) {
				return
			}
			response, err := s.Client().Do(r)
			if assert.Nil(t, err, err) {
				assert.Equal(t, response.StatusCode, http.StatusOK)

				assert.Equal(t, tt.want.method, gotMethod)
				assert.Equal(t, tt.want.url, gotUrl)
				assert.Equal(t, tt.want.header, gotHeader)
				assert.Equal(t, tt.want.body, gotBody)
			}
		})
	}
}
