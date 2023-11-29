package resterror

import "fmt"

// Wrapf 包装错误，添加消息。
func Wrapf(err error, format string, a ...any) error {
	return fmt.Errorf("%s, error: [%w]", fmt.Sprintf(format, a...), err)
}
