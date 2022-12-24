package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/wencan/fastrest/restcodecs/restmime"
	"github.com/wencan/fastrest/restcodecs/restvalues"
	"github.com/wencan/fastrest/resterror"
	"github.com/wencan/fastrest/restutils"
)

// RequestErrorWrapper 请求处理错误包装。可以用来包装或覆盖请求错误。
var RequestErrorWrapper = func(ctx context.Context, err error) error {
	return resterror.ErrorWithStatus(err, resterror.StatusInvalidArgument)
}

// ValidateErrorWrapper 请求校验错误包装。可以用来包装或覆盖请求错误。
var ValidateErrorWrapper = func(ctx context.Context, err error) error {
	return resterror.ErrorWithStatus(err, resterror.StatusInvalidArgument)
}

// ReadRequestFunc 解析请求的函数的签名。
type ReadRequestFunc func(ctx context.Context, dest interface{}, r *http.Request) error

// ReadRequest 解析请求到对象。
// 支持GET的查询参数、POST/PUT/PATCH的Content-Type为application/json、application/x-www-form-urlencoded、application/x-protobuf的请求实体。
// 解析GET查询参数和application/x-www-form-urlencoded实体，需要dest对象字段带schema标签。
func ReadRequest(ctx context.Context, dest interface{}, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		err := restvalues.Decoder.Decode(dest, r.URL.Query())
		if err != nil {
			return RequestErrorWrapper(ctx, err)
		}
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		defer r.Body.Close()
		contentType := r.Header.Get("Content-Type")
		err := restmime.Unmarshal(dest, contentType, r.Body)
		if err != nil {
			return RequestErrorWrapper(ctx, err)
		}
	default:
		return RequestErrorWrapper(ctx, errors.New("Unsupported method: "+r.Method))
	}

	return nil
}

// ReadValidateRequest 解析请求到对象。
// 会用github.com/go-playground/validator校验对象字段值。
func ReadValidateRequest(ctx context.Context, dest interface{}, r *http.Request) error {
	err := ReadRequest(ctx, dest, r)
	if err != nil {
		return err
	}

	err = restutils.ValidateStruct(ctx, dest)
	if err != nil {
		return ValidateErrorWrapper(ctx, err)
	}
	return nil
}
