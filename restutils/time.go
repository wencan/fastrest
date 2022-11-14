package restutils

import (
	"sync/atomic"
	"time"
)

// now 当前时间戳，精确到Millisecond。
var now int64

func init() {
	go func() {
		ticker := time.NewTicker(time.Millisecond * 100) // 0.1s
		for {
			c := <-ticker.C
			atomic.StoreInt64(&now, c.UnixMilli())
		}
	}()
}

// CoarseTimestamp 当前时间戳。单位秒。精确到0.1秒。
func CoarseTimestamp() float64 {
	s := atomic.LoadInt64(&now)
	if s == 0 {
		s = time.Now().UnixMilli()
	}

	return float64(s) / 1000
}
