package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRandomNumber(t *testing.T) {
	assert.Equal(t, 4, getRandomNumber())
}
