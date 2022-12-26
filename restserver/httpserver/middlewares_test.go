package httpserver

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestChainHandlerMiddlewares(t *testing.T) {

	makeMiddleware := func(name string) func(next HandleFunc) HandleFunc {
		return func(next HandleFunc) HandleFunc {
			return func(ctx context.Context, request interface{}) (response interface{}, err error) {
				tags := request.(*[]string)
				*tags = append(*tags, name+" start")
				response, err = next(ctx, request)
				*tags = append(*tags, name+" end")
				return response, err
			}
		}
	}
	middleware1 := makeMiddleware("middleware1")
	middleware2 := makeMiddleware("middleware2")
	middleware3 := makeMiddleware("middleware3")

	handler := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		tags := request.(*[]string)
		*tags = append(*tags, "work")
		return struct{}{}, nil
	}

	chain := ChainHandlerMiddlewares(middleware1, middleware2, middleware3)
	var tags []string
	chain(handler)(context.TODO(), &tags)

	wantTags := []string{"middleware1 start", "middleware2 start", "middleware3 start", "work", "middleware3 end", "middleware2 end", "middleware1 end"}
	if !reflect.DeepEqual(tags, wantTags) {
		t.Fatalf("want: %v, got: %v", strings.Join(wantTags, "; "), strings.Join(tags, "; "))
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	type args struct {
		handle HandleFunc
	}
	tests := []struct {
		name      string
		args      args
		wantError string
	}{
		{
			name: "panic",
			args: args{
				handle: func(ctx context.Context, request interface{}) (response interface{}, err error) {
					panic("test")
				},
			},
			wantError: "panic: test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handle := RecoveryMiddleware(tt.args.handle)
			_, err := handle(context.TODO(), nil)
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
