package lib

import (
	"strings"
	"testing"
)

// The scanner runs on every outgoing prompt, so its cost has to stay
// proportional to prompt size even when the input is adversarial: a long run
// of BIP-39 words forces a checksum validation at every offset.
func BenchmarkRedactSecretsWordlistFlood(b *testing.B) {
	flood := strings.TrimSpace(strings.Repeat("abandon ", 5000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RedactSecrets(flood, SecretRedactHybrid)
	}
}

func BenchmarkRedactSecretsTypicalPrompt(b *testing.B) {
	prompt := strings.Repeat("Explain how Morpheus session escrow works and why my balance looks wrong. ", 40)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RedactSecrets(prompt, SecretRedactHybrid)
	}
}
