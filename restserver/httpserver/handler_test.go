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

	"github.com/golang/mock/gomock"
	"github.com/wencan/fastrest/resterror"
	"github.com/wencan/fastrest/restserver/httpserver/mock_httpserver"
	"github.com/wencan/fastrest/restutils"
)

func Test_GetHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type Request struct {
		Greeting string `schema:"greeting"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	type args struct {
		url                string
		mockNewRequestFunc func() interface{}
		mockHandleFunc     func(ctx context.Context, request interface{}) (interface{}, error)
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
				url: "/echo?greeting=hello",
				mockNewRequestFunc: func() interface{} {
					return new(Request)
				},
				mockHandleFunc: func(ctx context.Context, request interface{}) (interface{}, error) {
					req := request.(*Request)
					return Response{
						Echo: req.Greeting,
					}, nil
				},
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
				url: "/echo?greeting=hello",
				mockNewRequestFunc: func() interface{} {
					return new(string)
				},
				mockHandleFunc: func(ctx context.Context, request interface{}) (interface{}, error) {
					return nil, resterror.ErrorWithStatus(errors.New("test"), resterror.StatusInvalidArgument)
				},
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
			handling := mock_httpserver.NewMockHandling(ctrl)
			handling.EXPECT().NewRequest().DoAndReturn(tt.args.mockNewRequestFunc).AnyTimes()
			handling.EXPECT().Handle(gomock.Any(), gomock.Any()).DoAndReturn(tt.args.mockHandleFunc).AnyTimes()

			hander := NewHandler(handling)
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
