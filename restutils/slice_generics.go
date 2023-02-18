//go:build go1.18
// +build go1.18

package restutils

// SliceContains 切片是否包含指定值。
func SliceContains[T comparable](s []T, a T) bool {
	for _, one := range s {
		if one == a {
			return true
		}
	}
	return false
}

// UnduplicatedSlice 返回一个新的元素不重复的切片，新切片保留元素顺序。
func UnduplicatedSlice[T comparable](s []T) []T {
	m := make(map[T]interface{}, len(s))
	n := make([]T, 0, len(s))
	for _, a := range s {
		_, exist := m[a]
		if exist {
			continue
		}

		n = append(n, a)
		m[a] = nil
	}
	return n
}
