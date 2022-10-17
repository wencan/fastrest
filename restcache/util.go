package restcache

import (
	"math/rand"
	"time"
)

// getTTL 计算出区间内的一个TTL值。
func getTTL(TTLRange [2]time.Duration) time.Duration {
	var r = rand.New(rand.NewSource(time.Now().UnixNano())) // rand不是并发安全的
	gap := TTLRange[1] - TTLRange[0]
	ttl := TTLRange[0] + time.Duration(r.Intn(int(gap.Seconds())))*time.Second

	return ttl
}
