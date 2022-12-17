package resterror

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Status 错误状态码。
type Status interface {
	HTTPStatusCode() int

	GRPCCode() codes.Code
}

// builtInStatus 内置的Status接口的实现。
type builtInStatus struct {
	HttpStatusCode int

	GRpcCode codes.Code
}

// HTTPStatusCode 实现Status接口。
func (s builtInStatus) HTTPStatusCode() int {
	return s.HttpStatusCode
}

// GRPCCode 实现Status接口。
func (s builtInStatus) GRPCCode() codes.Code {
	return s.GRpcCode
}

var (
	// StatusOk 成功。
	StatusOk Status = builtInStatus{
		HttpStatusCode: http.StatusOK,
		GRpcCode:       codes.OK,
	}

	// StatusInvalidArgument 无效参数。
	StatusInvalidArgument Status = builtInStatus{
		HttpStatusCode: http.StatusBadRequest,
		GRpcCode:       codes.InvalidArgument,
	}

	// StatusNotFound 没找到。
	StatusNotFound Status = builtInStatus{
		HttpStatusCode: http.StatusNotFound,
		GRpcCode:       codes.NotFound,
	}

	// StatusAlreadyExists 已经存在，冲突。
	StatusAlreadyExists Status = builtInStatus{
		HttpStatusCode: http.StatusConflict,
		GRpcCode:       codes.AlreadyExists,
	}

	// StatusPermissionDenied 拒绝访问。
	StatusPermissionDenied Status = builtInStatus{
		HttpStatusCode: http.StatusForbidden,
		GRpcCode:       codes.PermissionDenied,
	}

	// StatusFailedPrecondition 条件不满足
	StatusFailedPrecondition Status = builtInStatus{
		HttpStatusCode: http.StatusPreconditionFailed,
		GRpcCode:       codes.FailedPrecondition,
	}

	// StatusUnimplemented 未实现。
	StatusUnimplemented Status = builtInStatus{
		HttpStatusCode: http.StatusNotImplemented,
		GRpcCode:       codes.Unimplemented,
	}

	// StatusInternal 内部错误。
	StatusInternal Status = builtInStatus{
		HttpStatusCode: http.StatusInternalServerError,
		GRpcCode:       codes.Internal,
	}

	// StatusUnavailable 服务当前不可用。
	StatusUnavailable Status = builtInStatus{
		HttpStatusCode: http.StatusServiceUnavailable,
		GRpcCode:       codes.Unavailable,
	}

	// StatusUnauthenticated 未认证。
	StatusUnauthenticated Status = builtInStatus{
		HttpStatusCode: http.StatusUnauthorized,
		GRpcCode:       codes.Unauthenticated,
	}
)

// StatusError 带了错误状态码的error实现。
type StatusError struct {
	// error std错误
	error

	status Status
}

// ErrorWithStatus 包装状态码，返回一个新的error。
func ErrorWithStatus(err error, status Status) StatusError {
	return StatusError{
		error:  err,
		status: status,
	}
}

// HTTPStatusCode 实现github.com/wencan/fastrest/restserver/httpserver的错误接口HTTPStatusError。
func (statusError StatusError) HTTPStatusCode() int {
	return statusError.status.HTTPStatusCode()
}

// GRPCStatus 实现grpc-go的接口。
// google.golang.org/grpc/status.Code()可以取得StatusError的GRPCCode。
func (statusError StatusError) GRPCStatus() *status.Status {
	return status.New(statusError.status.GRPCCode(), statusError.error.Error())
}
