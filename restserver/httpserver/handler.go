package httpserver

import (
	"context"
	"net/http"
)

// HandlerFunc 需要开发者实现的服务处理函数的签名。HandlerFunc将会作为NewHandlerFunc的入参。
type HandlerFunc func(r *http.Request) (response interface{}, err error)

// NewStdHandlerFunc 创建http.HandlerFunc的函数的签名。
type NewStdHandlerFunc func(handler HandlerFunc) http.HandlerFunc

// HandlerFactoryConfig Handler工厂配置。配置全部可选。
type HandlerFactoryConfig struct {
	// RequestInterceptor 请求拦截器。
	RequestInterceptor func(r *http.Request) (overwriteRequest *http.Request, err error)

	// ResponseInterceptor 响应拦截器。
	ResponseInterceptor func(ctx context.Context, response interface{}, err error) (overwriteResponse interface{}, overwriteErr error)
}

// DefaultHandlerFactoryConfig 默认Handler工厂配置。可覆盖。
var DefaultHandlerFactoryConfig HandlerFactoryConfig

func newHandler(config *HandlerFactoryConfig, handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var response interface{}
		var err error

		if config.RequestInterceptor != nil {
			r, err = config.RequestInterceptor(r)
		}

		if err == nil {
			response, err = handler(r)
		}

		if config.ResponseInterceptor != nil {
			response, err = config.ResponseInterceptor(ctx, response, err)
		}

		statusCode := HTTPStatus(err)
		w.WriteHeader(statusCode)

		if response != nil {
			accept := r.Header.Get("Accept")
			WriteResponse(ctx, accept, response, w)
		}
	}
}

// NewHandlerFactory 创建一个创建http.HandlerFunc的工厂。
// 如果config为nil，使用DefaultHandlerFactoryConfig。
func NewHandlerFactory(config *HandlerFactoryConfig) NewStdHandlerFunc {
	if config == nil {
		config = &DefaultHandlerFactoryConfig
	}

	return func(handler HandlerFunc) http.HandlerFunc {
		return newHandler(config, handler)
	}
}

// NewHandler 使用DefaultHandlerFactoryConfig创建一个Handler。
func NewHandler(handler HandlerFunc) http.HandlerFunc {
	return newHandler(&DefaultHandlerFactoryConfig, handler)
}

func T(config *HandlerFactoryConfig, handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var response interface{}
		var err error

		statusCode := HTTPStatus(err)
		if response != nil {
			accept := r.Header.Get("Accept")
			WriteResponse(ctx, accept, response, w)
		}
		w.WriteHeader(statusCode)
	}
}
