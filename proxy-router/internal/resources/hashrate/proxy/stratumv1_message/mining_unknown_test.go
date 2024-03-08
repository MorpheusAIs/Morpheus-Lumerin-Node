package stratumv1_message

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiningUnknown(t *testing.T) {
	msg := []byte(`{"id":1,"method":"mining.unknown","params":[{"minimum-difficulty.value":2048,"version-rolling.mask":"1fffe000","version-rolling.min-bit-count":2}]}`)
	parsed, err := ParseGenericMessage(msg)
	if err != nil {
		t.FailNow()
	}
	msg2 := parsed.Serialize()

	// quick and dirty assuming the order and formatting of fields remains the same
	// TODO: write more reliable test
	require.Equal(t, string(msg), string(msg2))
}
