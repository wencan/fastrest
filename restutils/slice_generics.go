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
