package httpserver

import (
	"errors"
	"net/http"

	"github.com/wencan/fastrest/resterror"
)

// HTTPStatusError 提供Http状态码的错误接口。
type HTTPStatusError interface {
	error
	HTTPStatusCode() int
}

// HTTPStatus err对应的HTTP状态文本和状态码。
// 如果err为nil，返回200；
// 如果err实现了HTTPStatusError接口，返回HTTPStatus()的结果；
// 否则返回500。
func HTTPStatusCode(err error) int {
	err = resterror.FixNilError(err)
	if err == nil {
		return http.StatusOK
	}

	var statusError HTTPStatusError
	if errors.As(err, &statusError) {
		return statusError.HTTPStatusCode()
	}

	return http.StatusInternalServerError
}
