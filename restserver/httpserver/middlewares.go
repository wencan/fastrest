package httpserver

import (
	"context"
	"fmt"

	"github.com/wencan/fastrest/resterror"
)

// HandlerMiddleware Handler中间件。
type HandlerMiddleware func(next HandleFunc) HandleFunc

// HandlerMiddlewareChain 中间件链。
func ChainHandlerMiddlewares(middlewares ...HandlerMiddleware) HandlerMiddleware {
	return func(next HandleFunc) HandleFunc {
		if len(middlewares) == 0 {
			return next
		}

		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			current := next
			for i := len(middlewares) - 1; i >= 0; i-- {
				current = middlewares[i](current)
			}
			return current(ctx, request)
		}
	}
}

// HandleRecovery RecoveryMiddleware 的recovery处理函数。
// 默认处理，可覆盖。
var HandleRecovery = func(ctx context.Context, recovery interface{}) (overwriteRecovery interface{}) {
	return resterror.NewPanicError(recovery)
}

// RecoveryMiddleware 处理panic的中间件。recover()返回值将被转为error。
func RecoveryMiddleware(next HandleFunc) HandleFunc {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		defer func() {
			recovery := recover()
			if recovery != nil {
				recovery = HandleRecovery(ctx, recovery)
				err, _ = recovery.(error)
				if err == nil {
					err = fmt.Errorf("%v", recovery)
				}
			}
		}()
		response, err = next(ctx, request)
		return
	}
}
