//go:build go1.18
// +build go1.18

package restutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapKeys2Slice(t *testing.T) {
	m := map[int]string{
		1: "1",
		2: "2",
		3: "3",
	}
	keys := MapKeys2Slice(m)
	assert.Equalf(t, []int{1, 2, 3}, AscSorted(keys), "got keys: %v", keys)
}

func TestMapValues2Slice(t *testing.T) {
	m := map[int]string{
		1: "1",
		2: "2",
		3: "3",
	}
	values := MapValues2Slice(m)
	assert.Equalf(t, []string{"1", "2", "3"}, AscSorted(values), "got values: %v", values)
}
