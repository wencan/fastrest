package httpserver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/wencan/fastrest/resterror"
	"github.com/wencan/fastrest/restutils"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

func Test_GetHandler(t *testing.T) {
	type Request struct {
		Greeting string `schema:"greeting"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	type args struct {
		handler HandlerFunc
		url     string
	}
	type want struct {
		statusCode   int
		header       http.Header
		responseBody []byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "test_get",
			args: args{
				handler: func(r *http.Request) (response interface{}, err error) {
					var req Request
					err = ReadRequest(r.Context(), &req, r)
					if err != nil {
						return nil, err
					}
					return Response{
						Echo: req.Greeting,
					}, nil
				},
				url: "/echo?greeting=hello",
			},
			want: want{
				statusCode: http.StatusOK,
				header: http.Header{
					"Content-Length": []string{strconv.Itoa(len("{\"echo\":\"hello\"}\n"))},
					"Content-Type":   []string{"application/json"},
				},
				responseBody: []byte("{\"echo\":\"hello\"}\n"),
			},
		},
		{
			name: "test_get_400",
			args: args{
				handler: func(r *http.Request) (response interface{}, err error) {
					return nil, resterror.ErrorWithStatus(errors.New("test"), resterror.StatusInvalidArgument)
				},
				url: "/echo?greeting=hello",
			},
			want: want{
				statusCode: http.StatusBadRequest,
				header: http.Header{
					"Content-Length": []string{"0"},
				},
				responseBody: []byte(``),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hander := NewHandler(tt.args.handler)
			s := httptest.NewServer(hander)
			defer s.Close()

			client := s.Client()
			resp, err := client.Get(s.URL + tt.args.url)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != tt.want.statusCode {
				t.Fatalf("want status code: %d, got status code: %d", tt.want.statusCode, resp.StatusCode)
			}

			gotHeader := make(http.Header)
			for key, values := range resp.Header {
				if restutils.StringSliceContains([]string{"Date"}, key) {
					continue
				}
				gotHeader[key] = values
			}
			if !reflect.DeepEqual(gotHeader, tt.want.header) {
				t.Fatalf("want header: %v, got header: %v", tt.want.header, gotHeader)
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if !bytes.Equal(data, tt.want.responseBody) {
				t.Fatalf("want response: %v, got response: %v", tt.want.responseBody, data)
			}
		})
	}
}

// testGRPCServer is used to implement helloworld.GreeterServer.
type testGRPCServer struct {
	helloworld.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s testGRPCServer) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func TestNewHandlerFunc(t *testing.T) {
	type Request struct {
		Greeting string `schema:"greeting" json:"greeting"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	type args struct {
		f               interface{}
		readRequestFunc ReadRequestFunc

		method string
		url    string
		header http.Header
		body   io.Reader
	}
	type want struct {
		statusCode   int
		header       http.Header
		responseBody []byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "test_get_request-response",
			args: args{
				f: func(ctx context.Context, req *Request) (response Response, err error) {
					return Response{
						Echo: req.Greeting,
					}, nil
				},
				method: http.MethodGet,
				url:    "/echo?greeting=hello",
			},
			want: want{
				statusCode: http.StatusOK,
				header: http.Header{
					"Content-Length": []string{strconv.Itoa(len("{\"echo\":\"hello\"}\n"))},
					"Content-Type":   []string{"application/json"},
				},
				responseBody: []byte("{\"echo\":\"hello\"}\n"),
			},
		},
		{
			name: "test_get_request-any",
			args: args{
				f: func(ctx context.Context, req *Request) (response interface{}, err error) {
					return Response{
						Echo: req.Greeting,
					}, nil
				},
				method: http.MethodGet,
				url:    "/echo?greeting=hello",
			},
			want: want{
				statusCode: http.StatusOK,
				header: http.Header{
					"Content-Length": []string{strconv.Itoa(len("{\"echo\":\"hello\"}\n"))},
					"Content-Type":   []string{"application/json"},
				},
				responseBody: []byte("{\"echo\":\"hello\"}\n"),
			},
		},
		{
			name: "test_from-grpc-method",
			args: args{
				f:      testGRPCServer{}.SayHello,
				method: http.MethodPost,
				url:    "/echo",
				header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				body: bytes.NewBufferString("{\"name\":\"Tom\"}"),
			},
			want: want{
				statusCode: http.StatusOK,
				header: http.Header{
					"Content-Length": []string{strconv.Itoa(len("{\"message\":\"Hello Tom\"}\n"))},
					"Content-Type":   []string{"application/json"},
				},
				responseBody: []byte("{\"message\":\"Hello Tom\"}\n"),
			},
		},
		{
			name: "test_get_400",
			args: args{
				f: func(ctx context.Context, req *Request) (response interface{}, err error) {
					return nil, resterror.ErrorWithStatus(errors.New("test"), resterror.StatusInvalidArgument)
				},
				method: http.MethodGet,
				url:    "/echo?greeting=hello",
			},
			want: want{
				statusCode: http.StatusBadRequest,
				header: http.Header{
					"Content-Length": []string{"0"},
				},
				responseBody: []byte(``),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hander := NewHandler(NewHandlerFunc(tt.args.f, tt.args.readRequestFunc))
			s := httptest.NewServer(hander)
			defer s.Close()

			client := s.Client()
			r, err := http.NewRequestWithContext(context.TODO(), tt.args.method, s.URL+tt.args.url, tt.args.body)
			if err != nil {
				t.Fatal(err)
				return
			}
			for key, values := range tt.args.header {
				for _, value := range values {
					r.Header.Add(key, value)
				}
			}
			resp, err := client.Do(r)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want.statusCode {
				t.Fatalf("want status code: %d, got status code: %d", tt.want.statusCode, resp.StatusCode)
			}

			gotHeader := make(http.Header)
			for key, values := range resp.Header {
				if restutils.StringSliceContains([]string{"Date"}, key) {
					continue
				}
				gotHeader[key] = values
			}
			if !reflect.DeepEqual(gotHeader, tt.want.header) {
				t.Fatalf("want header: %v, got header: %v", tt.want.header, gotHeader)
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if !bytes.Equal(tt.want.responseBody, data) {
				t.Fatalf("want response: %v, got response: %v", tt.want.responseBody, data)
			}
		})
	}
}
