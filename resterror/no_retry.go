package resterror

import (
	"errors"
	"fmt"
)

// NoRetryError 不重试的错误接口。
type NoRetryError interface {
	error
	NoRetry()
}

// IsNoRetry 判断是否是一个NoRetry错误。
func IsNoRetry(err error) bool {
	if err == nil {
		return false
	}
	var e noRetryError
	return errors.As(err, &e)
}

type noRetryError struct {
	error
}

// NoRetry 实现NoRetryError。
func (err noRetryError) NoRetry() {}

// WrapNoRetryError 包装一个error，返回一个NoRetryError的实现实例。
func WrapNoRetryError(err error) error {
	return noRetryError{error: err}
}

// FormatNoRetryError 格式化一个NoRetryError的实现实例。
func FormatNoRetryError(format string, a ...interface{}) error {
	return noRetryError{error: fmt.Errorf(format, a...)}
}
