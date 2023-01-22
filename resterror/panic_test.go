package resterror

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoWithRecover(t *testing.T) {
	err := Recover(func() {
		panic("test")
	})
	assert.NotNil(t, err)

	err = Recover(func() {})
	assert.Nil(t, err)
}
