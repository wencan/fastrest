//go:build go1.18
// +build go1.18

package restutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceContains(t *testing.T) {
	assert.True(t, SliceContains([]int{1, 2, 3}, int(1)))
	assert.False(t, SliceContains([]int{1, 2, 3}, int(0)))

	assert.True(t, SliceContains([]string{"1", "2", "3"}, string("1")))
	assert.False(t, SliceContains([]string{"1", "2", "3"}, string("0")))
}

func TestUnduplicatedSlice(t *testing.T) {
	nums := []int{2, 2, 3, 3, 6, 7, 7, 4, 2, 0, 0}
	assert.Equal(t, []int{2, 3, 6, 7, 4, 0}, UnduplicatedSlice(nums))
}
