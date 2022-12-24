package httpserver

import (
	"context"
	"log"
	"net/http"
)

// Handling 处理逻辑的接口。GenericsHandling实现了该接口。
type Handling interface {
	// NewRequest 创建请求对象，返回请求对象的地址/指针。
	NewRequest() interface{}

	// Handle 处理请求，返回响应。
	Handle(ctx context.Context, request interface{}) (response interface{}, err error)
}

// HandleFunc Handling的Handle方法的类型。
type HandleFunc func(ctx context.Context, request interface{}) (response interface{}, err error)

// DefaultHandlerFactory 默认的Handler工厂。可修改，可覆盖。
var DefaultHandlerFactory = HandlerFactory{
	ReadRequestFunc:   ReadRequest,
	Middleware:        RecoveryMiddleware,
	WriteResponseFunc: WriteResponse,
}

// HandlerFactory Handler工厂。
type HandlerFactory struct {
	// ReadRequestFunc 读请求函数。默认是：ReadRequest。
	ReadRequestFunc ReadRequestFunc

	// Middleware 中间件。多个中间件可以用ChainHandlerMiddlewares串联起来。默认是：RecoveryMiddleware。
	Middleware HandlerMiddleware

	// WriteResponseFunc 写响应的函数的签名。默认是：WriteResponse。
	WriteResponseFunc WriteResponseFunc
}

// NewHandler 创建一个http.Handler。
func (factory HandlerFactory) NewHandler(handling Handling) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		ctx := r.Context()
		var request, response interface{}
		var err error
		var handle HandleFunc

		// 加上中间件
		if factory.Middleware != nil {
			handle = factory.Middleware(handling.Handle)
		} else {
			handle = handling.Handle
		}

		request = handling.NewRequest()                // new请求对象
		err = factory.ReadRequestFunc(ctx, request, r) // 读请求
		if err == nil {
			response, err = handle(ctx, request) // 处理
		}

		err = factory.WriteResponseFunc(ctx, w, r, response, err) // 写响应
		if err != nil {
			log.Printf("failed to write response, error: %s\n", err)
		}
	}
}

// NewHandler 基于DefaultHandlerFactory创建一个http.Handler。
func NewHandler(handling Handling) http.HandlerFunc {
	return DefaultHandlerFactory.NewHandler(handling)
}
