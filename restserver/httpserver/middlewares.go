package httpserver

import (
	"fmt"
	"net/http"

	"github.com/wencan/fastrest/restutils"
)

// HandlerMiddleware Handler中间件。
type HandlerMiddleware func(next HandlerFunc) HandlerFunc

// HandlerMiddlewareChain 中间件链。
func ChainHandlerMiddlewares(middlewares ...HandlerMiddleware) HandlerMiddleware {
	return func(next HandlerFunc) HandlerFunc {
		if len(middlewares) == 0 {
			return next
		}

		return func(r *http.Request) (response interface{}, err error) {
			current := next
			for i := len(middlewares) - 1; i >= 0; i-- {
				current = middlewares[i](current)
			}
			return current(r)
		}
	}
}

// HandleRecovery RecoveryMiddleware 的recovery处理函数。
// 默认处理，可覆盖。
var HandleRecovery = func(r *http.Request, recovery interface{}) (overwriteRecovery interface{}) {
	stack := restutils.CaptureStack(2)
	return fmt.Errorf("painc: %v\n%s", recovery, stack.StackTrace())
}

// RecoveryMiddleware 处理panic的中间件。
func RecoveryMiddleware(next HandlerFunc) HandlerFunc {
	return func(r *http.Request) (response interface{}, err error) {
		defer func() {
			recovery := recover()
			if recovery != nil {
				recovery = HandleRecovery(r, recovery)
				err, _ = recovery.(error)
				if err == nil {
					err = fmt.Errorf("%v", recovery)
				}
			}
		}()
		response, err = next(r)
		return
	}
}
