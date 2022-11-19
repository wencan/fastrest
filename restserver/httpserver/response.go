package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/wencan/fastrest/restcodecs/restmime"
)

// WriteResponseFunc 输出响应的函数的签名。
type WriteResponseFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request, response interface{}, err error) error

// WriteResponse 将响应体写出。
// response将被转为响应实体。
// 响应Content-Type根据请求的Accept推导。
// 如果err非nil，尝试转为HTTPStatusError接口，获取错误码。
func WriteResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, response interface{}, err error) error {
	statusCode := http.StatusOK
	if err != nil {
		statusCode = HTTPStatusCode(err)
	}

	// 先写header
	// 再状态码
	// 再body

	if response == nil {
		w.WriteHeader(statusCode)
		return nil
	}

	accept := r.Header.Get("Accept")
	if accept == "" {
		accept = "*/*"
	}
	// 根据Accept要求，找出返回的content type。
	contentType := restmime.AcceptableMarshalContentType(accept)
	if contentType == "" {
		return fmt.Errorf("invalid Accept: [%s]", accept)
	}
	// 先设置header
	w.Header().Set("Content-Type", contentType)
	// 再输出状态码和header
	w.WriteHeader(statusCode)
	// 最后输出body
	err = restmime.Marshal(response, contentType, w)
	if err != nil {
		return err
	}

	return nil
}
