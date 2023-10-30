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

// HitIndexes 根据未命中的索引，得到命中的索引
func HitIndexes[T any](collection []T, missIndexes []int) []int {
	var missPos int
	var hitIndexes []int
	for idx := range collection {
		if missPos >= len(missIndexes) {
			hitIndexes = append(hitIndexes, idx)
			continue
		}

		if missIndexes[missPos] == idx {
			// miss
			missPos += 1
		} else {
			// hit
			hitIndexes = append(hitIndexes, idx)
		}
	}
	return hitIndexes
}
