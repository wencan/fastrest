package httpclient

import (
	"context"
	"net/http"

	"github.com/wencan/fastrest/restcodecs/restmime"
)

// ReadResponseBody 反序列化body到dest对象。不会close Body。
// 如果body为空，会报错。
func ReadResponseBody(ctx context.Context, dest interface{}, response *http.Response) error {
	contentType := response.Header.Get("Content-Type")
	return restmime.Unmarshal(dest, contentType, response.Body)
}
