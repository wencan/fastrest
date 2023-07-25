package restcache

import (
	"math/rand"
	"time"
)

// getTTL 计算出区间内的一个TTL值。
func getTTL(TTLRange [2]time.Duration) time.Duration {
	gap := TTLRange[1] - TTLRange[0]
	if gap == 0 {
		return TTLRange[0]
	}

	ttl := TTLRange[0] + time.Duration(rand.Intn(int(gap.Seconds())))*time.Second

	return ttl
}
