package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/wencan/fastrest/restcodecs/restmime"
	"github.com/wencan/fastrest/restcodecs/restvalues"
	"github.com/wencan/fastrest/resterror"
	"github.com/wencan/fastrest/restutils"
	"google.golang.org/grpc/codes"
)

// RequestErrorWrapper 请求处理错误包装。可以用来包装或覆盖请求错误。
var RequestErrorWrapper = func(ctx context.Context, err error) error {
	return resterror.ErrorWithStatus(err, http.StatusBadRequest, codes.InvalidArgument)
}

// ReadRequest 解析请求到对象。支持GET的请求参数、POST/PUT/PATCH的Content-Type为application/json和application/x-www-form-urlencoded的请求实体。
// 解析GET请求参数，需要dest对象字段带schema标签。
func ReadRequest(ctx context.Context, dest interface{}, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		err := restvalues.Decoder.Decode(dest, r.URL.Query())
		if err != nil {
			return RequestErrorWrapper(ctx, err)
		}
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		defer r.Body.Close()
		accept := r.Header.Get("Content-Type")
		err := restmime.Unmarshal(dest, accept, r.Body)
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
		return RequestErrorWrapper(ctx, err)
	}
	return nil
}
