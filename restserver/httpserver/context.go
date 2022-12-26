package httpserver

import (
	"context"
	"net/http"
)

var contextKeyRequest struct{}

func newContextWithRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, contextKeyRequest, r)
}

// RequestFromContext 从上下文中取得*http.Request对象。
func RequestFromContext(ctx context.Context) *http.Request {
	r, _ := ctx.Value(contextKeyRequest).(*http.Request)
	return r
}
