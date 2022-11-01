package httpserver

import (
	"context"
	"log"
	"net/http"
	"reflect"
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

// NewReflectHandler 基于配置，创建一个http.Handler。
// f 的签名必须是： func (context.Context, <RequestType>) (<ResponseType>, error)。
// readRequest读请求函数默认为ReadRequest。
// NewReflectHandler会利用反射，创建请求对象，解析请求数据到请求对象，调用函数f，输出响应。
func (config HandlerConfig) NewReflectHandler(f interface{}, readRequestFunc ReadRequestFunc) http.HandlerFunc {
	fValue := reflect.ValueOf(f)
	fType := fValue.Type()
	if !fType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		panic("the first parameter of the Handler must be context")
	}
	if !fType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		panic("the second return value of the Handler must be error")
	}
	requestArgsType := fType.In(1)
	for requestArgsType.Kind() == reflect.Ptr {
		requestArgsType = requestArgsType.Elem()
	}

	if readRequestFunc == nil {
		readRequestFunc = ReadRequest
	}

	return config.NewHandler(func(r *http.Request) (response interface{}, err error) {
		requestValue := reflect.New(requestArgsType)
		err = readRequestFunc(r.Context(), requestValue.Interface(), r)
		if err != nil {
			return nil, err
		}

		ins := []reflect.Value{reflect.ValueOf(r.Context()), requestValue}
		outs := fValue.Call(ins)
		responseValue := outs[0]
		errValue := outs[1]

		response = responseValue.Interface()
		if !errValue.IsNil() {
			err = errValue.Interface().(error)
		}

		return response, err
	})
}

// NewHandler 使用DefaultHandlerConfig创建一个http.Handler。
func NewHandler(handler HandlerFunc) http.HandlerFunc {
	return DefaultHandlerConfig.NewHandler(handler)
}

// NewReflectHandler 使用DefaultHandlerConfig创建一个http.Handler。
// f 的签名必须是： func (context.Context, <RequestType>) (<ResponseType>, error)。
// readRequest读请求函数，默认为ReadRequest。
// NewReflectHandler会利用反射，New请求对象，解析请求数据到请求对象，调用函数f，输出响应。
func NewReflectHandler(f interface{}, readRequestFunc ReadRequestFunc) http.HandlerFunc {
	return DefaultHandlerConfig.NewReflectHandler(f, readRequestFunc)
}
