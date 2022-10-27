package restcache

import (
	"math/rand"
	"time"
)

var randSource = rand.NewSource(time.Now().UnixNano())

// getTTL 计算出区间内的一个TTL值。
func getTTL(TTLRange [2]time.Duration) time.Duration {
	gap := TTLRange[1] - TTLRange[0]
	if gap == 0 {
		return TTLRange[0]
	}

	var r = rand.New(randSource) // rand不是并发安全的
	ttl := TTLRange[0] + time.Duration(r.Intn(int(gap.Seconds())))*time.Second

	return ttl
}
