package httpserver

import (
	"context"
	"net/http"

	"github.com/wencan/fastrest/restcodecs/restmime"
)

// WriteResponse 将响应写出。
func WriteResponse(ctx context.Context, contentType string, response interface{}, w http.ResponseWriter) error {
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	return restmime.Marshal(response, contentType, w)
}
