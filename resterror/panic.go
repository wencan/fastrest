package resterror

import (
	"fmt"
	"io"

	"github.com/wencan/fastrest/restutils"
)

var _ error = PanicError{}
var _ error = &PanicError{}

// PanicError panic错误。
type PanicError struct {
	// recovery Recover()的结果。
	recovery interface{}

	// stack recover时的调用栈。
	stack restutils.Stack
}

// NewPanicError 创建一个PanicError。
func NewPanicError(recovery interface{}) PanicError {
	return PanicError{
		recovery: recovery,
		stack:    restutils.CaptureStack(2),
	}
}

// String 实现error。返回字符串格式为：panic <recovery>。
func (err PanicError) Error() string {
	return "panic: " + fmt.Sprint(err.recovery)
}

// Format 实现fmt.Formatter。
func (err PanicError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('#'), s.Flag('+'):
			io.WriteString(s, "panic: "+fmt.Sprint(err.recovery))
			io.WriteString(s, "\n")
			io.WriteString(s, err.stack.StackTrace())
		default:
			io.WriteString(s, "panic: "+fmt.Sprint(err.recovery))
		}
	default:
		io.WriteString(s, "panic: "+fmt.Sprint(err.recovery))
	}
}

// StackTrace 调用栈。
func (err PanicError) StackTrace() string {
	return err.stack.StackTrace()
}

// Recover 执行函数f。如果函数发生panic，返回一个PanicError错误；否则返回nil。
func Recover(f func()) (panicErr *PanicError) {
	defer func() {
		recovery := recover()
		if recovery != nil {
			err := NewPanicError(recovery)
			panicErr = &err
		}
	}()

	f()
	return nil
}
