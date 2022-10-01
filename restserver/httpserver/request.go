package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/wencan/fastrest/restcodecs/restvalues"
	"github.com/wencan/fastrest/restutils"
)

// ReadRequest 解析请求到对象。支持GET的请求参数、POST/PUT/PATCH的Content-Type为application/json和application/x-www-form-urlencoded的请求实体。
// 解析GET请求参数，需要dest对象字段带schema标签。
func ReadRequest(ctx context.Context, dest interface{}, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		err := restvalues.Decoder.Decode(dest, r.URL.Query())
		if err != nil {
			return err
		}
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		defer r.Body.Close()
		accept := r.Header.Get("Accept")
		err := restutils.UnmarshalContent(dest, accept, r.Body)
		if err != nil {
			return err
		}
	default:
		return errors.New("Unsupported method: " + r.Method)
	}

	return nil
}

// ReadValidateRequest 解析请求到对象，会用。
func ReadValidateRequest(ctx context.Context, dest interface{}, r *http.Request) error {
	err := ReadRequest(ctx, dest, r)
	if err != nil {
		return err
	}

	return restutils.ValidateStruct(ctx, dest)
}
