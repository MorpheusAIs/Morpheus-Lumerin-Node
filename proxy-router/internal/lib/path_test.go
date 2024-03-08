package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeFilename(t *testing.T) {
	src := `2023-03-03T13:57:52Z-!!!@#$^&{}[]^&*()\/~|-stratum.slushpool.com:3333_[::]:123.log`
	exp := "2023-03-03t13_57_52z-_____________________-stratum.slushpool.com_3333______123.log"
	res := SanitizeFilename(src)
	assert.Equal(t, res, exp)
}
