package validator

import (
	"crypto/rand"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatorDiff(t *testing.T) {
	msg := GetTestMsg()

	diff, ok := ValidateDiff(msg.xnonce, uint(msg.xnonce2size), uint64(msg.diff), msg.vmask, msg.notify, msg.submit1)

	require.Truef(t, ok, "Result diff (%d) doesn't meet difficulty target (%.2f)", diff, msg.diff)
}

func TestValidatorDiffInvalidMsg(t *testing.T) {}

func TestValidatorDiffMalformedMsg(t *testing.T) {}

func BenchmarkValidateDiff(b *testing.B) {
	msg := GetTestMsg()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ValidateDiff(msg.xnonce, uint(msg.xnonce2size), uint64(msg.diff), msg.vmask, msg.notify, msg.submit1)
	}
}

func BenchmarkSha256Std(b *testing.B) {
	buf := make([]byte, 260)
	_, err := rand.Read(buf)
	require.NoError(b, err)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sha256.Sum256(buf)
	}
}
