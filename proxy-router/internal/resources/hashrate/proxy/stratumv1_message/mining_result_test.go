package stratumv1_message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKIk(t *testing.T) {
	ID := 1
	msg := NewMiningResultJobNotFound(ID)

	msg, err := ParseMiningResult(msg.Serialize())

	assert.NoError(t, err)
	assert.Equal(t, msg.ID, ID)
	assert.True(t, msg.IsError())
	assert.Equal(t, msg.GetError(), `["21","Job not found"]`)
}
