package httpserver

import (
	"context"
	"log"
	"net/http"
	"reflect"
)

// HandlerFunc 需要开发者实现的服务处理函数的签名。HandlerFunc将会作为NewHandlerFunc的入参。
type HandlerFunc func(r *http.Request) (response interface{}, err error)

// DefaultHandlerFactory 默认的Handler工厂。可修改，可覆盖。
var DefaultHandlerFactory = HandlerFactory{
	ReadRequestFunc:   ReadRequest,
	DefaultAccept:     "*/*",
	Middleware:        RecoveryMiddleware,
	WriteResponseFunc: WriteResponse,
}

// HandlerFactory Handler工厂。
type HandlerFactory struct {
	// ReadRequestFunc 读请求函数。默认是：ReadRequest。
	ReadRequestFunc ReadRequestFunc

	// DefaultAccept 请求Header Accept的缺省值。默认是：*/*。
	DefaultAccept string

	// Middleware 中间件。多个中间件可以用ChainHandlerMiddlewares串联起来。默认是：RecoveryMiddleware。
	Middleware HandlerMiddleware

	// WriteResponseFunc 写响应的函数的签名。默认是：WriteResponse。
	WriteResponseFunc WriteResponseFunc
}

// NewHandler 创建一个http.Handler。
func (factory HandlerFactory) NewHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		ctx := r.Context()
		var response interface{}
		var err error

		if factory.Middleware != nil {
			handler = factory.Middleware(handler)
		}

		response, err = handler(r)

		if factory.DefaultAccept != "" {
			if _, ok := r.Header["Accept"]; !ok {
				r = r.Clone(ctx)
				r.Header.Set("Accept", factory.DefaultAccept)
			}
		}

		err = factory.WriteResponseFunc(ctx, w, r, response, err)
		if err != nil {
			log.Printf("failed to write response, error: %s\n", err)
		}
	}
}

// NewHandlerFunc 通过反射，将一个处理函数转换为HandlerFunc。
// f 的签名必须是： func (context.Context, <RequestType>) (<ResponseType>, error)。
// readRequest用于解析请求，默认为ReadRequest。
// NewHandler + NewHandlerFunc 可以将一个gRPC服务方法，转为http.HandlerFunc。
// 如果f是一个gRPC方法，方法返回的错误码需要同时支持gRPC和HTTP。可以用resterror.ErrorWithStatus包装错误，或者编写一个实现了httpserver.HTTPStatusError接口和GRPCStatus() *google.golang.org/grpc/status.Status方法的错误结构。
func NewHandlerFunc(f interface{}, readRequestFunc ReadRequestFunc) HandlerFunc {
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

	return func(r *http.Request) (response interface{}, err error) {
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
	}
}

// NewHandlerFunc 通过反射，将一个处理函数转换为HandlerFunc。
// f 的签名必须是： func (context.Context, <RequestType>) (<ResponseType>, error)。
// NewHandler + NewHandlerFunc 可以将一个gRPC服务方法，转为http.HandlerFunc。
// 如果f是一个gRPC方法，方法返回的错误码需要同时支持gRPC和HTTP。可以用resterror.ErrorWithStatus包装错误，或者编写一个实现了httpserver.HTTPStatusError接口和GRPCStatus() *google.golang.org/grpc/status.Status方法的错误结构。
func (factory HandlerFactory) NewHandlerFunc(f interface{}) HandlerFunc {
	return NewHandlerFunc(f, factory.ReadRequestFunc)
}

// NewReflectHandler NewHandler + NewHandlerFunc的快速方法。
func (factory HandlerFactory) NewReflectHandler(f interface{}) http.HandlerFunc {
	handler := factory.NewHandlerFunc(f)
	return factory.NewHandler(handler)
}

// NewHandler 基于DefaultHandlerFactory创建一个http.Handler。
func NewHandler(handler HandlerFunc) http.HandlerFunc {
	return DefaultHandlerFactory.NewHandler(handler)
}

// NewReflectHandler 基于DefaultHandlerFactory的NewHandler + NewHandlerFunc的快速方法。。
func NewReflectHandler(f interface{}) http.HandlerFunc {
	return DefaultHandlerFactory.NewReflectHandler(f)
}
