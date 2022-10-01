package resterror

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StatusError 带了协议状态的error实现。
type StatusError struct {
	// error std错误
	error

	// HTTPStatusCode http状态码
	HTTPStatusCode int

	// GRPCCode grpc状态码
	GRPCCode codes.Code
}

// ErrorWithStatus 包装协议状态码，返回一个新的error。
func ErrorWithStatus(err error, httpStatusCode int, grpcCode codes.Code) StatusError {
	return StatusError{
		error:          err,
		HTTPStatusCode: httpStatusCode,
		GRPCCode:       grpcCode,
	}
}

// HTTPStatus 实现github.com/wencan/fastrest/restserver/httpserver的错误接口HTTPStatusError。
func (statusError StatusError) HTTPStatus() int {
	return statusError.HTTPStatusCode
}

// GRPCStatus 实现grpc-go的接口。
// google.golang.org/grpc/status.Code()可以取得ErrorWithStatus的GRPCCode。
func (statusError StatusError) GRPCStatus() *status.Status {
	return status.New(statusError.GRPCCode, statusError.error.Error())
}
