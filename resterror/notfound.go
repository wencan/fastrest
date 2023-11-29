package resterror

import (
	"errors"
	"fmt"
)

// NotFoundError 没找到的错误接口。
type NotFoundError interface {
	error
	NotFound()
}

// IsNotFound 判断是否是一个NotFound错误。
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var e notFoundError
	return errors.As(err, &e)
}

type notFoundError struct {
	error
}

// NotFound 实现NotFoundError。
func (err notFoundError) NotFound() {}

// WrapNotFoundError 包装一个error，返回一个NotFoundError的实现实例。
func WrapNotFoundError(err error) error {
	return notFoundError{error: err}
}

// FormatNotFoundError 格式化一个NotFoundError的实现实例。
func FormatNotFoundError(format string, a ...interface{}) error {
	return notFoundError{error: fmt.Errorf(format, a...)}
}
