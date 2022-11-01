package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/wencan/fastrest/restcodecs/restmime"
)

// WriteResponse 将响应体写出。
// 如果accept为空，表示没content type的要求。
func WriteResponse(ctx context.Context, statusCode int, accept string, response interface{}, w http.ResponseWriter) error {
	if response == nil {
		w.WriteHeader(statusCode)
		return nil
	}

	if accept == "" {
		accept = "*/*"
	}

	contentType := restmime.AcceptableMarshalContentType(accept)
	if contentType == "" {
		return fmt.Errorf("invalid Accept: [%s]", accept)
	}

	// 先设置header
	w.Header().Set("Content-Type", contentType)

	// 再输出状态码和header
	w.WriteHeader(statusCode)

	// 最后输出body
	err := restmime.Marshal(response, contentType, w)
	if err != nil {
		return err
	}

	return nil
}
