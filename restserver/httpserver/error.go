package httpserver

import (
	"net/http"

	"github.com/wencan/fastrest/resterror"
)

// HTTPStatusError 提供Http状态码的错误接口。
type HTTPStatusError interface {
	error
	HTTPStatus() int
}

// HTTPStatus err对应的财务码。
// 如果err为nil，返回200；
// 如果err实现了HTTPStatusError接口，返回HTTPStatus()的结果；
// 否则返回500。
func HTTPStatus(err error) int {
	err = resterror.FixNilError(err)
	if err == nil {
		return http.StatusOK
	}

	statusError, _ := err.(HTTPStatusError)
	if statusError != nil {
		return statusError.HTTPStatus()
	}

	return http.StatusInternalServerError
}
