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

func TestAsPanic(t *testing.T) {
	_, isPanic := AsPanic(nil)
	assert.False(t, isPanic)

	_, isPanic = AsPanic(NewPanicError(struct{}{}))
	assert.True(t, isPanic)
}
