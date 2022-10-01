package httpserver

import (
	"context"
	"net/http"

	"github.com/wencan/fastrest/restcodecs/restjson"
)

// WriteResponse 将响应写出。
// 暂时只会输出为json。
func WriteResponse(ctx context.Context, accept string, response interface{}, w http.ResponseWriter) error {
	encoder := restjson.NewEncoder(w)
	return encoder.Encode(response)
}
