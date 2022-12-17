package httpclient

import (
	"fmt"
	"net/http"

	"github.com/wencan/fastrest/resterror"
)

var statusCode2ErrorStatusMap = map[int]resterror.Status{
	http.StatusOK:                  resterror.StatusOk,
	http.StatusBadRequest:          resterror.StatusInvalidArgument,
	http.StatusNotFound:            resterror.StatusNotFound,
	http.StatusConflict:            resterror.StatusAlreadyExists,
	http.StatusForbidden:           resterror.StatusPermissionDenied,
	http.StatusPreconditionFailed:  resterror.StatusFailedPrecondition,
	http.StatusNotImplemented:      resterror.StatusUnimplemented,
	http.StatusInternalServerError: resterror.StatusInternal,
	http.StatusServiceUnavailable:  resterror.StatusUnavailable,
	http.StatusUnauthorized:        resterror.StatusUnauthenticated,
}

// StatusCodeError HTTP状态码转为带状态的错误。
func StatusCodeError(statusCode int, format string, a ...interface{}) resterror.StatusError {
	status, ok := statusCode2ErrorStatusMap[statusCode]
	if !ok {
		status = resterror.StatusInternal
	}

	return resterror.ErrorWithStatus(fmt.Errorf(format, a...), status)
}
