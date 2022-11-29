package httpclient

import (
	"context"
	"net/http"

	"github.com/wencan/fastrest/restcodecs/restmime"
)

// ReadResponseFunc 读请求函数签名。
type ReadResponseFunc func(ctx context.Context, dest interface{}, response *http.Response) error

// ReadResponseBody 反序列化body到dest对象。
// 不会检查状态码、Content-Length，不会close Body。
// 如果body为空，会报错。
func ReadResponseBody(ctx context.Context, dest interface{}, response *http.Response) error {
	contentType := response.Header.Get("Content-Type")
	return restmime.Unmarshal(dest, contentType, response.Body)
}

// ReadResponse 解析请求。对于非200的状态码，返回错误。解析实体。不会close Body。
func ReadResponse(ctx context.Context, dest interface{}, response *http.Response) error {
	if response.StatusCode != http.StatusOK {
		return StatusCodeError(response.StatusCode, "upstream server error")
	}

	contentLength := response.Header.Get("Content-Length")
	if contentLength != "" && contentLength != "0" {
		return ReadResponseBody(ctx, dest, response)
	}

	return nil
}
