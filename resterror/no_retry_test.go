package resterror

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNoRetry(t *testing.T) {
	assert.False(t, IsNoRetry(nil))

	err := FormatNoRetryError("no retry")
	assert.True(t, IsNoRetry(err))

	err = WrapNoRetryError(fmt.Errorf("no retry"))
	assert.True(t, IsNoRetry(err))

	err = fmt.Errorf("no retry")
	assert.False(t, IsNoRetry(err))
}
