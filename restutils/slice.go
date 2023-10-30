package restutils

// IntSliceContains []int是否包含一个指定的int值。
func IntSliceContains(ints []int, n int) bool {
	for _, one := range ints {
		if one == n {
			return true
		}
	}
	return false
}

// StringSliceContains []string是否包含一个指定的string值。
func StringSliceContains(strs []string, str string) bool {
	for _, one := range strs {
		if one == str {
			return true
		}
	}
	return false
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
