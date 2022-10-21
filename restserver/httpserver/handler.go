package httpserver

import (
	"log"
	"net/http"
)

// HandlerFunc 需要开发者实现的服务处理函数的签名。HandlerFunc将会作为NewHandlerFunc的入参。
type HandlerFunc func(r *http.Request) (response interface{}, err error)

// HandlerConfig Handler配置。配置全部可选。
type HandlerConfig struct {
	// Middleware 中间件。如果需要多个中间件，可以用ChainHandlerMiddlewares创建的中间件链。
	Middleware HandlerMiddleware

	// DefaultAccept 保底的Accept。
	// 如果请求没有通过Header Accept指定，则使用该值。
	DefaultAccept string
}

// DefaultHandlerConfig 默认Handler配置。可覆盖。
var DefaultHandlerConfig HandlerConfig

// NewHandler 基于配置，创建一个http.Handler。
func (config HandlerConfig) NewHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var response interface{}
		var err error

		if config.Middleware != nil {
			handler = config.Middleware(handler)
		}
		response, err = handler(r)

		statusCode := HTTPStatusCode(err)
		w.WriteHeader(statusCode)

		if response != nil {
			accept := r.Header.Get("Accept")
			if accept == "" {
				accept = config.DefaultAccept
			}
			err = WriteResponse(ctx, accept, response, w)
			if err != nil {
				log.Printf("failed to write response, error: %s\n", err)
			}
		}
	}
}

// NewHandler 使用DefaultHandlerConfig创建一个http.Handler。
func NewHandler(handler HandlerFunc) http.HandlerFunc {
	return DefaultHandlerConfig.NewHandler(handler)
}
