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
