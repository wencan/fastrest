package resterror

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNotFound(t *testing.T) {
	assert.False(t, IsNotFound(nil))

	err := FormatNotFoundError("not found")
	assert.True(t, IsNotFound(err))

	err = WrapNotFoundError(fmt.Errorf("not found"))
	assert.True(t, IsNotFound(err))

	err = fmt.Errorf("not found")
	assert.False(t, IsNotFound(err))
}
