package restutils

import (
	"bytes"
	"fmt"
	"runtime"
)

// Stack 调用栈。
type Stack struct {
	Frames []runtime.Frame
}

// CaptureStack 捕获调用栈。
// skip为0表示栈从CaptureStack开始，为1表示栈从CaptureStack的调用者开始。
func CaptureStack(skip int) Stack {
	var pcs [512]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	if n == 0 {
		return Stack{}
	}
	frames := runtime.CallersFrames(pcs[:n])

	var stack Stack
	for {
		frame, more := frames.Next()
		// if strings.Contains(frame.File, "/src/runtime/") {
		// 	break
		// }
		stack.Frames = append(stack.Frames, frame)

		if !more {
			break
		}
	}
	return stack
}

// StackTrace 格式化为多行字符串的堆栈跟踪信息。
func (stack Stack) StackTrace() string {
	var buffer bytes.Buffer
	for _, frame := range stack.Frames {
		fmt.Fprintf(&buffer, "%s(...)\n", frame.Function)
		fmt.Fprintf(&buffer, "    %s:%d\n", frame.File, frame.Line)
	}
	return buffer.String()
}

// String 实现fmt.Stringer。
func (stack Stack) String() string {
	if len(stack.Frames) == 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d", stack.Frames[0].File, stack.Frames[0].Line)
}
