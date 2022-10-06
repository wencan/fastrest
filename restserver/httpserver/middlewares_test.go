package httpserver

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestChainHandlerMiddlewares(t *testing.T) {
	var tags []string

	makeMiddleware := func(name string) func(next HandlerFunc) HandlerFunc {
		return func(next HandlerFunc) HandlerFunc {
			return func(r *http.Request) (response interface{}, err error) {
				tags = append(tags, name+" start")
				response, err = next(r)
				tags = append(tags, name+" end")
				return response, err
			}
		}
	}
	middleware1 := makeMiddleware("middleware1")
	middleware2 := makeMiddleware("middleware2")
	middleware3 := makeMiddleware("middleware3")

	handler := func(r *http.Request) (response interface{}, err error) {
		tags = append(tags, "work")
		return struct{}{}, nil
	}

	chain := ChainHandlerMiddlewares(middleware1, middleware2, middleware3)
	chain(handler)(httptest.NewRequest(http.MethodGet, "/", nil))

	wantTags := []string{"middleware1 start", "middleware2 start", "middleware3 start", "work", "middleware3 end", "middleware2 end", "middleware1 end"}
	if !reflect.DeepEqual(tags, wantTags) {
		t.Fatalf("want: %v, got: %v", strings.Join(wantTags, "; "), strings.Join(tags, "; "))
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	type args struct {
		handler HandlerFunc
	}
	tests := []struct {
		name      string
		args      args
		wantError string
	}{
		{
			name: "panic",
			args: args{
				handler: func(r *http.Request) (response interface{}, err error) {
					panic("test")
				},
			},
			wantError: "panic: test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RecoveryMiddleware(tt.args.handler)
			_, err := handler(httptest.NewRequest(http.MethodGet, "/test", nil))
			var gotError string
			if err != nil {
				gotError = err.Error()
			}
			if gotError != tt.wantError {
				t.Errorf("want error: [%s], got error: [%s]", tt.wantError, gotError)
			}
		})
	}
}
