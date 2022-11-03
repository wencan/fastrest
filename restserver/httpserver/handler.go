package httpserver

import (
	"context"
	"log"
	"net/http"
	"reflect"
)

// HandlerFunc 需要开发者实现的服务处理函数的签名。HandlerFunc将会作为NewHandlerFunc的入参。
type HandlerFunc func(r *http.Request) (response interface{}, err error)

// NewHandler 创建一个http.Handler。
func NewHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		ctx := r.Context()
		var response interface{}
		var err error

		response, err = handler(r)

		err = WriteResponse(ctx, w, r, response, err)
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
