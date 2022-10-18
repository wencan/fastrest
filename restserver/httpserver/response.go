package httpserver

import (
	"context"
	"net/http"

	"github.com/wencan/fastrest/restcodecs/restmime"
)

// WriteResponse 将响应写出。
// 暂时只会输出为json。
func WriteResponse(ctx context.Context, accept string, response interface{}, w http.ResponseWriter) error {
	return restmime.Marshal(response, accept, w)
}
