package lib

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapErrorInheritance(t *testing.T) {
	grandparent := errors.New("grandparent error")
	parent := errors.New("parent error")
	child := errors.New("child error")

	err := WrapError(grandparent, WrapError(parent, child))

	assert.True(t, errors.Is(err, grandparent))
	assert.True(t, errors.Is(err, parent))
	assert.True(t, errors.Is(err, child))
}

func TestErrorMessage(t *testing.T) {
	grandparent := errors.New("grandparent error")
	parent := errors.New("parent error")
	child := errors.New("child error")

	actErrMsg := WrapError(grandparent, WrapError(parent, child)).Error()
	expErrMsg := fmt.Sprintf("%s: %s: %s", grandparent.Error(), parent.Error(), child.Error())

	assert.Equal(t, expErrMsg, actErrMsg)
}
