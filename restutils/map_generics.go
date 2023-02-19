//go:build go1.18
// +build go1.18

package restutils

// MapKeys2Slice 提取map的key组成切片返回。
func MapKeys2Slice[K comparable, V any](m map[K]V) []K {
	s := make([]K, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}

// MapValues2Slice 提取map的value组成切片返回。
func MapValues2Slice[K comparable, V any](m map[K]V) []V {
	s := make([]V, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	return s
}
