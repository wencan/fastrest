//go:build go1.18
// +build go1.18

package restutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAscSorted(t *testing.T) {
	nums := []int{6, 3, 4, 2, 1, 5, 4}
	assert.Equal(t, []int{1, 2, 3, 4, 4, 5, 6}, AscSorted(nums))

	strings := []string{"6", "3", "4", "2", "1", "5", "4"}
	assert.Equal(t, []string{"1", "2", "3", "4", "4", "5", "6"}, AscSorted(strings))
}

func TestDescSorted(t *testing.T) {
	nums := []int{6, 3, 4, 2, 1, 5, 4}
	assert.Equal(t, []int{6, 5, 4, 4, 3, 2, 1}, DescSorted(nums))

	strings := []string{"6", "3", "4", "2", "1", "5", "4"}
	assert.Equal(t, []string{"6", "5", "4", "4", "3", "2", "1"}, DescSorted(strings))
}
