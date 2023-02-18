//go:build go1.18
// +build go1.18

package restutils

import (
	"sort"

	"golang.org/x/exp/constraints"
)

type sortable[T constraints.Ordered] []T

func (s sortable[T]) Len() int {
	return len(s)
}

func (s sortable[T]) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortable[T]) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// AscSorted 返回升序排列的新切片。
func AscSorted[T constraints.Ordered](s []T) []T {
	n := make([]T, len(s))
	copy(n, s)

	sort.Sort(sortable[T](n))
	return n
}

// DescSorted 返回降序排列的新切片。
func DescSorted[T constraints.Ordered](s []T) []T {
	n := make([]T, len(s))
	copy(n, s)

	sort.Sort(sort.Reverse(sortable[T](n)))
	return n
}
