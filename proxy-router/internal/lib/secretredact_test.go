package lib

import (
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

const (
	// 64 hex chars, the shape of a secp256k1 private key and also of a tx
	// hash, a block hash, and a Morpheus session ID.
	hex64  = "4c0883a69102937d6231471b5dbb6204fe512961708279e2f1f0a9d2c1b5c8ea"
	hex64b = "8f2a559490e9b0a7c8b4d3e1f60572849bc3d7a1e5f09b2c6d48a3715e0c9fb2"

	// BIP-39 test vectors: all-zero entropy at 128 and 256 bits.
	validMnemonic12 = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	validMnemonic24 = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon " +
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"

	// Twelve wordlist words whose checksum does not validate.
	badChecksum12 = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"
)

func TestRedactSecretsHex(t *testing.T) {
	tests := []struct {
		name string
		mode SecretRedactMode
		in   string
		want string
		n    int
	}{
		{
			name: "bare 64-hex is a key in hybrid",
			mode: SecretRedactHybrid,
			in:   "import " + hex64 + " please",
			want: "import " + RedactedPrivateKey + " please",
			n:    1,
		},
		{
			name: "0x-prefixed without a cue is left alone in hybrid",
			mode: SecretRedactHybrid,
			in:   "what happened in tx 0x" + hex64 + "?",
			want: "what happened in tx 0x" + hex64 + "?",
			n:    0,
		},
		{
			name: "0x-prefixed with a preceding cue is redacted in hybrid",
			mode: SecretRedactHybrid,
			in:   "my private key is 0x" + hex64,
			want: "my private key is " + RedactedPrivateKey,
			n:    1,
		},
		{
			name: "0x-prefixed with a trailing cue is redacted in hybrid",
			mode: SecretRedactHybrid,
			in:   "0x" + hex64 + " is the wallet key, right?",
			want: RedactedPrivateKey + " is the wallet key, right?",
			n:    1,
		},
		{
			name: "env-var assignment is redacted in hybrid",
			mode: SecretRedactHybrid,
			in:   "PRIVATE_KEY=0x" + hex64,
			want: "PRIVATE_KEY=" + RedactedPrivateKey,
			n:    1,
		},
		{
			name: "strict redacts 0x-prefixed with no cue at all",
			mode: SecretRedactStrict,
			in:   "what happened in tx 0x" + hex64 + "?",
			want: "what happened in tx " + RedactedPrivateKey + "?",
			n:    1,
		},
		{
			name: "off leaves everything",
			mode: SecretRedactOff,
			in:   "my private key is 0x" + hex64,
			want: "my private key is 0x" + hex64,
			n:    0,
		},
		{
			name: "two adjacent bare keys are both redacted",
			mode: SecretRedactHybrid,
			in:   hex64 + " " + hex64b,
			want: RedactedPrivateKey + " " + RedactedPrivateKey,
			n:    2,
		},
		{
			name: "a 20-byte address is not a key",
			mode: SecretRedactStrict,
			in:   "send it to 0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
			want: "send it to 0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
			n:    0,
		},
		{
			name: "a longer hex run is not a key",
			mode: SecretRedactStrict,
			in:   "sig 0x" + hex64 + hex64b,
			want: "sig 0x" + hex64 + hex64b,
			n:    0,
		},
		{
			name: "63 hex chars is not a key",
			mode: SecretRedactStrict,
			in:   "value " + hex64[:63],
			want: "value " + hex64[:63],
			n:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, n := RedactSecrets(tt.in, tt.mode)
			if got != tt.want {
				t.Errorf("text\n got: %q\nwant: %q", got, tt.want)
			}
			if n != tt.n {
				t.Errorf("count: got %d, want %d", n, tt.n)
			}
		})
	}
}

func TestRedactSecretsSeedPhrase(t *testing.T) {
	tests := []struct {
		name string
		mode SecretRedactMode
		in   string
		want string
		n    int
	}{
		{
			name: "valid 12-word phrase",
			mode: SecretRedactHybrid,
			in:   "restore with " + validMnemonic12 + " thanks",
			want: "restore with " + RedactedSeedPhrase + " thanks",
			n:    1,
		},
		{
			name: "valid 24-word phrase prefers the longer match",
			mode: SecretRedactHybrid,
			in:   validMnemonic24,
			want: RedactedSeedPhrase,
			n:    1,
		},
		{
			name: "numbered and newline separated phrase",
			mode: SecretRedactHybrid,
			in:   "1. " + strings.ReplaceAll(validMnemonic12, " ", "\n2. "),
			want: "1. " + RedactedSeedPhrase,
			n:    1,
		},
		{
			name: "mixed case phrase",
			mode: SecretRedactHybrid,
			in:   strings.ToUpper(validMnemonic12),
			want: RedactedSeedPhrase,
			n:    1,
		},
		{
			name: "bad checksum survives hybrid",
			mode: SecretRedactHybrid,
			in:   badChecksum12,
			want: badChecksum12,
			n:    0,
		},
		{
			name: "bad checksum is caught by strict",
			mode: SecretRedactStrict,
			in:   badChecksum12,
			want: RedactedSeedPhrase,
			n:    1,
		},
		{
			name: "ordinary prose is untouched",
			mode: SecretRedactStrict,
			in:   "Please explain how session escrow works on the Morpheus marketplace and why my balance looks wrong.",
			want: "Please explain how session escrow works on the Morpheus marketplace and why my balance looks wrong.",
			n:    0,
		},
		{
			name: "off leaves the phrase",
			mode: SecretRedactOff,
			in:   validMnemonic12,
			want: validMnemonic12,
			n:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, n := RedactSecrets(tt.in, tt.mode)
			if got != tt.want {
				t.Errorf("text\n got: %q\nwant: %q", got, tt.want)
			}
			if n != tt.n {
				t.Errorf("count: got %d, want %d", n, tt.n)
			}
		})
	}
}

// A phrase broken by a paragraph gap must not be stitched back into one match.
func TestRedactSecretsSeedPhraseGapBreaksRun(t *testing.T) {
	words := strings.Fields(validMnemonic12)
	in := strings.Join(words[:6], " ") + "\n\nand then some unrelated text here\n\n" + strings.Join(words[6:], " ")
	got, n := RedactSecrets(in, SecretRedactHybrid)
	if n != 0 || got != in {
		t.Errorf("split phrase should not match: got %q (n=%d)", got, n)
	}
}

func TestRedactSecretsBothKinds(t *testing.T) {
	in := "key " + hex64 + " and phrase " + validMnemonic12
	want := "key " + RedactedPrivateKey + " and phrase " + RedactedSeedPhrase
	got, n := RedactSecrets(in, SecretRedactHybrid)
	if got != want {
		t.Errorf("text\n got: %q\nwant: %q", got, want)
	}
	if n != 2 {
		t.Errorf("count: got %d, want 2", n)
	}
}

func TestRedactMessagesSecrets(t *testing.T) {
	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: "You are a helpful assistant."},
		{Role: openai.ChatMessageRoleUser, Content: "sweep this wallet: " + hex64},
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{Type: openai.ChatMessagePartTypeText, Text: "seed is " + validMnemonic12},
				{Type: openai.ChatMessagePartTypeImageURL, ImageURL: &openai.ChatMessageImageURL{URL: "https://example.com/a.png"}},
			},
		},
		{
			Role: openai.ChatMessageRoleAssistant,
			ToolCalls: []openai.ToolCall{{
				Type:     openai.ToolTypeFunction,
				Function: openai.FunctionCall{Name: "send", Arguments: `{"key":"` + hex64b + `"}`},
			}},
		},
	}

	n := RedactMessagesSecrets(msgs, SecretRedactHybrid)
	if n != 3 {
		t.Fatalf("count: got %d, want 3", n)
	}
	if msgs[0].Content != "You are a helpful assistant." {
		t.Errorf("system message was modified: %q", msgs[0].Content)
	}
	if msgs[1].Content != "sweep this wallet: "+RedactedPrivateKey {
		t.Errorf("user content: %q", msgs[1].Content)
	}
	if msgs[2].MultiContent[0].Text != "seed is "+RedactedSeedPhrase {
		t.Errorf("multicontent text: %q", msgs[2].MultiContent[0].Text)
	}
	if msgs[2].MultiContent[1].ImageURL.URL != "https://example.com/a.png" {
		t.Errorf("image url was modified: %q", msgs[2].MultiContent[1].ImageURL.URL)
	}
	if want := `{"key":"` + RedactedPrivateKey + `"}`; msgs[3].ToolCalls[0].Function.Arguments != want {
		t.Errorf("tool arguments: %q", msgs[3].ToolCalls[0].Function.Arguments)
	}
}

func TestRedactMessagesSecretsOffIsNoop(t *testing.T) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: hex64}}
	if n := RedactMessagesSecrets(msgs, SecretRedactOff); n != 0 {
		t.Fatalf("count: got %d, want 0", n)
	}
	if msgs[0].Content != hex64 {
		t.Errorf("content was modified: %q", msgs[0].Content)
	}
}

func TestParseSecretRedactMode(t *testing.T) {
	tests := map[string]SecretRedactMode{
		"off":       SecretRedactOff,
		"OFF":       SecretRedactOff,
		"  strict ": SecretRedactStrict,
		"hybrid":    SecretRedactHybrid,
		// Unset or misspelled must fail closed, not open.
		"":         SecretRedactHybrid,
		"nonsense": SecretRedactHybrid,
	}
	for in, want := range tests {
		if got := ParseSecretRedactMode(in); got != want {
			t.Errorf("ParseSecretRedactMode(%q) = %q, want %q", in, got, want)
		}
	}
}
