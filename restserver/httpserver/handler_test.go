package httpserver

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wencan/fastrest/resterror"
	"github.com/wencan/fastrest/restutils"
)

func Test_GetHandler(t *testing.T) {
	type Request struct {
		Greeting string `schema:"greeting"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	type args struct {
		config  *HandlerConfig
		handler HandlerFunc
		url     string
	}
	type want struct {
		statusCode   int
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
				config: &DefaultHandlerConfig,
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
				statusCode:   http.StatusOK,
				responseBody: []byte(`{"echo":"hello"}`),
			},
		},
		{
			name: "test_get_400",
			args: args{
				config: &DefaultHandlerConfig,
				handler: func(r *http.Request) (response interface{}, err error) {
					return nil, resterror.ErrorWithStatus(errors.New("test"), resterror.StatusInvalidArgument)
				},
				url: "/echo?greeting=hello",
			},
			want: want{
				statusCode:   http.StatusBadRequest,
				responseBody: []byte(``),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hander := tt.args.config.NewHandler(tt.args.handler)
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
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if restutils.CompareHumanizeString(string(tt.want.responseBody), string(data)) != 0 {
				t.Fatalf("want response: %s, got response: %s", string(tt.want.responseBody), string(data))
			}
		})
	}
}

func TestNewReflectHandler(t *testing.T) {
	type Request struct {
		Greeting string `schema:"greeting"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	type args struct {
		config          *HandlerConfig
		f               interface{}
		readRequestFunc ReadRequestFunc
		url             string
	}
	type want struct {
		statusCode   int
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
				config: &DefaultHandlerConfig,
				f: func(ctx context.Context, req *Request) (response Response, err error) {
					return Response{
						Echo: req.Greeting,
					}, nil
				},
				url: "/echo?greeting=hello",
			},
			want: want{
				statusCode:   http.StatusOK,
				responseBody: []byte(`{"echo":"hello"}`),
			},
		},
		{
			name: "test_get_request-any",
			args: args{
				config: &DefaultHandlerConfig,
				f: func(ctx context.Context, req *Request) (response interface{}, err error) {
					return Response{
						Echo: req.Greeting,
					}, nil
				},
				url: "/echo?greeting=hello",
			},
			want: want{
				statusCode:   http.StatusOK,
				responseBody: []byte(`{"echo":"hello"}`),
			},
		},
		{
			name: "test_get_400",
			args: args{
				config: &DefaultHandlerConfig,
				f: func(ctx context.Context, req *Request) (response interface{}, err error) {
					return nil, resterror.ErrorWithStatus(errors.New("test"), resterror.StatusInvalidArgument)
				},
				url: "/echo?greeting=hello",
			},
			want: want{
				statusCode:   http.StatusBadRequest,
				responseBody: []byte(``),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hander := tt.args.config.NewReflectHandler(tt.args.f, tt.args.readRequestFunc)
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
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if restutils.CompareHumanizeString(string(tt.want.responseBody), string(data)) != 0 {
				t.Fatalf("want response: %s, got response: %s", string(tt.want.responseBody), string(data))
			}
		})
	}
}
