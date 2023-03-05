package httpclient

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

func TestGet(t *testing.T) {
	largeData := make([]byte, 1024)
	_, err := rand.Read(largeData)
	if err != nil {
		panic(err)
	}
	oneK := base64.RawStdEncoding.EncodeToString(largeData)

	largeData = make([]byte, 1024*1024)
	_, err = rand.Read(largeData)
	if err != nil {
		panic(err)
	}
	oneM := base64.RawStdEncoding.EncodeToString(largeData)

	largeData = make([]byte, 1024*1014*10)
	_, err = rand.Read(largeData)
	if err != nil {
		panic(err)
	}
	tenM := base64.RawStdEncoding.EncodeToString(largeData)

	type args struct {
		dest  interface{}
		url   string
		query interface{}

		handlerFunc http.HandlerFunc
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "get_json_without_query",
			args: args{
				dest: &struct {
					Greeting string `json:"greeting"`
					Name     string `json:"name"`
				}{},
				url:   "/123/get_json_without_query",
				query: nil,
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Content-Type", restmime.MimeTypeJson)
					w.Write([]byte("{\"greeting\":\"hi\",\"name\":\"Tom\"}\n"))
				},
			},
			want: &struct {
				Greeting string `json:"greeting"`
				Name     string `json:"name"`
			}{
				Greeting: "hi",
				Name:     "Tom",
			},
		},
		{
			name: "get_json_with_query",
			args: args{
				dest: &struct {
					Greeting string `json:"greeting"`
					Name     string `json:"name"`
				}{},
				url:   "/123/get_json_with_query",
				query: url.Values{"greeting": []string{"hi"}, "name": []string{"張三"}},
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					greeting := r.URL.Query().Get("greeting")
					name := r.URL.Query().Get("name")

					w.Header().Add("Content-Type", restmime.MimeTypeJson)
					w.Write([]byte(fmt.Sprintf("{\"greeting\":\"%s\",\"name\":\"%s\"}\n", greeting, name)))
				},
			},
			want: &struct {
				Greeting string `json:"greeting"`
				Name     string `json:"name"`
			}{
				Greeting: "hi",
				Name:     "張三",
			},
		},
		{
			name: "get_json_with_schemaquery",
			args: args{
				dest: &struct {
					Greeting string `json:"greeting"`
					Name     string `json:"name"`
				}{},
				url: "/123/get_json_with_schemaquery",
				query: struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{
					Greeting: "hi",
					Name:     "王五",
				},
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					greeting := r.URL.Query().Get("greeting")
					name := r.URL.Query().Get("name")

					w.Header().Add("Content-Type", restmime.MimeTypeJson)
					w.Write([]byte(fmt.Sprintf("{\"greeting\":\"%s\",\"name\":\"%s\"}\n", greeting, name)))
				},
			},
			want: &struct {
				Greeting string `json:"greeting"`
				Name     string `json:"name"`
			}{
				Greeting: "hi",
				Name:     "王五",
			},
		},
		{
			name: "get_json_no_responseBody",
			args: args{
				dest:  nil,
				url:   "/123/get_json_no_responseBody",
				query: nil,
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				},
			},
		},
		{
			name: "get_protobuf_with_query",
			args: args{
				dest:  &pb.HelloReply{},
				url:   "/123/get_json_without_query",
				query: url.Values{"name": []string{"王五"}},
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Content-Type", restmime.MimeTypeProtobuf)
					data, _ := proto.Marshal(&pb.HelloReply{
						Message: r.URL.Query().Get("name"),
					})
					w.Write(data)
				},
			},
			want: &pb.HelloReply{Message: "王五"},
		},
		{
			name: "get_1k",
			args: args{
				dest: &struct {
					Data string `json:"data"`
				}{},
				url:   "/123/get_1k",
				query: nil,
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Content-Type", restmime.MimeTypeJson)
					w.Write([]byte("{\"data\":\"" + oneK + "\"}\n"))
				},
			},
			want: &struct {
				Data string `json:"data"`
			}{
				Data: oneK,
			},
		},
		{
			name: "get_1m",
			args: args{
				dest: &struct {
					Data string `json:"data"`
				}{},
				url:   "/123/get_1m",
				query: nil,
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Content-Type", restmime.MimeTypeJson)
					_, err := w.Write([]byte("{\"data\":\"" + oneM + "\"}\n"))
					if err != nil {
						panic(err)
					}
				},
			},
			want: &struct {
				Data string `json:"data"`
			}{
				Data: oneM,
			},
		},
		{
			name: "get_10m",
			args: args{
				dest: &struct {
					Data string `json:"data"`
				}{},
				url:   "/123/get_10m",
				query: nil,
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Content-Type", restmime.MimeTypeJson)
					_, err := w.Write([]byte("{\"data\":\"" + tenM + "\"}\n"))
					if err != nil {
						panic(err)
					}
				},
			},
			want: &struct {
				Data string `json:"data"`
			}{
				Data: tenM,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(tt.args.handlerFunc)
			defer s.Close()

			err := Get(context.TODO(), tt.args.dest, s.URL+tt.args.url, tt.args.query)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				if !assert.Nil(t, err) {
					return
				}

				wantMessage, _ := tt.want.(proto.Message)
				haveMessage, _ := tt.args.dest.(proto.Message)
				if wantMessage != nil && haveMessage != nil {
					assert.True(t, proto.Equal(wantMessage, haveMessage))
				} else {
					assert.Equal(t, tt.want, tt.args.dest)
				}
			}
		})
	}
}

func TestPost(t *testing.T) {
	type Request struct {
		Greeting string `schema:"greeting" json:"greeting"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	type args struct {
		dest        interface{}
		url         string
		contentType string
		body        interface{}

		handlerFunc http.HandlerFunc
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "post_json",
			args: args{
				dest:        &Response{},
				url:         "/post_json",
				contentType: restmime.MimeTypeJson,
				body: Request{
					Greeting: "hi",
				},
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					var request Request
					err := json.NewDecoder(r.Body).Decode(&request)
					r.Body.Close()
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					w.Header().Set("Content-Type", restmime.MimeTypeJson)
					w.Write([]byte(fmt.Sprintf("{\"echo\":\"%s\"}\n", request.Greeting)))
				},
			},
			want: &Response{Echo: "hi"},
		},
		{
			name: "post_protobuf",
			args: args{
				dest:        &pb.HelloReply{},
				url:         "/post_protobuf",
				contentType: restmime.MimeTypeProtobuf,
				body: &pb.HelloRequest{
					Name: "奥巴马",
				},
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					data, err := io.ReadAll(r.Body)
					r.Body.Close()
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					var request pb.HelloRequest
					err = proto.Unmarshal(data, &request)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					w.Header().Add("Content-Type", restmime.MimeTypeProtobuf)
					data, _ = proto.Marshal(&pb.HelloReply{
						Message: request.Name,
					})
					w.Write(data)
				},
			},
			want: &pb.HelloReply{
				Message: "奥巴马",
			},
		},
		{
			name: "post_from",
			args: args{
				dest:        &Response{},
				url:         "/post_from",
				contentType: restmime.MimeTypeForm,
				body: Request{
					Greeting: "普金",
				},
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					greeting := r.FormValue("greeting")

					w.Header().Set("Content-Type", restmime.MimeTypeJson)
					w.Write([]byte(fmt.Sprintf("{\"echo\":\"%s\"}\n", greeting)))
				},
			},
			want: &Response{Echo: "普金"},
		},
		{
			name: "post_noRequestBody",
			args: args{
				dest: &Response{},
				url:  "/post_noRequestBody",
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", restmime.MimeTypeJson)
					w.Write([]byte(fmt.Sprintf("{\"echo\":\"%s\"}\n", "noBody")))
				},
			},
			want: &Response{Echo: "noBody"},
		},
		{
			name: "post_noResponseBody",
			args: args{
				url:         "/post_noResponseBody",
				contentType: restmime.MimeTypeJson,
				body: Request{
					Greeting: "hi",
				},
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(tt.args.handlerFunc)
			defer s.Close()

			err := Post(context.TODO(), tt.args.dest, s.URL+tt.args.url, tt.args.contentType, tt.args.body)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				if !assert.Nil(t, err) {
					return
				}

				wantMessage, _ := tt.want.(proto.Message)
				haveMessage, _ := tt.args.dest.(proto.Message)
				if wantMessage != nil && haveMessage != nil {
					assert.True(t, proto.Equal(wantMessage, haveMessage))
				} else {
					assert.Equal(t, tt.want, tt.args.dest)
				}
			}
		})
	}
}
