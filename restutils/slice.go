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
