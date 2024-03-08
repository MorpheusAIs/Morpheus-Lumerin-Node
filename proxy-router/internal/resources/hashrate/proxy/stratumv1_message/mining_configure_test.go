package stratumv1_message

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiningConfigure(t *testing.T) {
	msg := NewMiningConfigure(1, &MiningConfigureExtensionParams{
		VersionRollingMask:        "00000000",
		VersionRollingMinBitCount: 2,
	})

	msgParsed, err := ParseStratumMessage(msg.Serialize())
	require.NoError(t, err)

	m, b := msgParsed.(*MiningConfigure).GetVersionRolling()
	require.Equal(t, "00000000", m)
	require.Equal(t, 2, b)
}
