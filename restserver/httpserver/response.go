package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/wencan/fastrest/restcodecs/restmime"
)

// WriteResponse 将响应写出。
// 如果accept为空，表示没content type的要求。
func WriteResponse(ctx context.Context, accept string, response interface{}, w http.ResponseWriter) error {
	if accept == "" {
		accept = "*/*"
	}

	contentType := restmime.AcceptableMarshalContentType(accept)
	if contentType == "" {
		return fmt.Errorf("invalid Accept: [%s]", accept)
	}

	w.Header().Set("Content-Type", contentType)
	err := restmime.Marshal(response, contentType, w)
	if err != nil {
		return err
	}
	return nil
}
