package lib

import (
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/tyler-smith/go-bip39"
)

// Consumer-side scrubbing of wallet secrets out of outgoing prompts.
//
// A prompt typed into MorpheusUI (or any OpenAI-compatible client pointed at
// this node) leaves the machine and lands on an independent provider's
// hardware. If the user pasted a private key or a seed phrase into that
// prompt, the key is now in a stranger's request log. Redaction happens here,
// at the consumer boundary, before the request is signed and shipped, so the
// scrubbed text is also what gets written to local chat history.
//
// The hard part is that an Ethereum private key is 32 random bytes, and so is
// a tx hash, a block hash, and a Morpheus session ID. Nothing in the shape of
// the string distinguishes them, so the mode below trades false positives
// (redacting a tx hash the user wanted the model to read) against false
// negatives (shipping a real key). SecretRedactHybrid is the default and
// leans on the one signal that does correlate in practice: keys get pasted
// bare, hashes get pasted with the 0x prefix.

// SecretRedactMode selects how aggressively outgoing prompts are scrubbed.
type SecretRedactMode string

const (
	// SecretRedactOff ships prompts through untouched.
	SecretRedactOff SecretRedactMode = "off"

	// SecretRedactHybrid redacts bare 64-hex blobs (no 0x prefix) always, and
	// 0x-prefixed ones only when a private-key cue sits nearby. Seed phrases
	// must pass the BIP-39 checksum. Pasted tx hashes and session IDs survive.
	SecretRedactHybrid SecretRedactMode = "hybrid"

	// SecretRedactStrict redacts every 64-hex blob regardless of prefix or
	// context, and every run of 12+ BIP-39 words whether or not the checksum
	// validates. No key escapes; tx hashes and session IDs do not survive.
	SecretRedactStrict SecretRedactMode = "strict"
)

// Placeholders are deliberately readable: the user should be able to tell from
// the model's reply that their secret never reached it.
const (
	RedactedPrivateKey = "[redacted:private-key]"
	RedactedSeedPhrase = "[redacted:seed-phrase]"
)

// cueWindow is how many bytes on either side of a 0x-prefixed hex blob get
// searched for a private-key cue in hybrid mode. Wide enough to cover
// "my private key is 0x..." and "0x... (that's the wallet key)", narrow enough
// that an unrelated mention earlier in a long prompt doesn't drag a tx hash in.
const cueWindow = 120

// maxMnemonicGap bounds the separator between two consecutive BIP-39 words.
// Seed phrases get pasted space-separated, newline-separated, and numbered
// ("1. abandon\n2. ability"), all of which fit; a paragraph break between two
// coincidental wordlist hits does not.
const maxMnemonicGap = 8

// bip39WordCounts are the legal BIP-39 phrase lengths, longest first so the
// scanner prefers the largest phrase it can prove at a given offset.
var bip39WordCounts = [...]int{24, 21, 18, 15, 12}

// minBip39Words is the shortest legal phrase, and therefore the shortest run
// worth scanning.
const minBip39Words = 12

var (
	// \b on both ends keeps this from firing on a fragment of a longer hex run
	// (a 128-hex signature, say), since RE2 finds no word boundary mid-run.
	hexSecretPattern = regexp.MustCompile(`\b(0[xX])?[0-9A-Fa-f]{64}\b`)

	privateKeyCuePattern = regexp.MustCompile(
		`(?i)(private[ _\-]?key|priv[ _\-]?key|secret[ _\-]?key|sign(?:ing|er)[ _\-]?key|wallet[ _\-]?key|keystore|seed[ _\-]?phrase|mnemonic|recovery[ _\-]?phrase)`,
	)

	wordPattern = regexp.MustCompile(`[A-Za-z]+`)
)

// bip39Words is the English wordlist as a set. Built once; the list is fixed
// at 2048 entries and shared with the HD wallet code, so a membership test is
// the cheap pre-filter before the much more expensive checksum validation.
var bip39Words = func() map[string]struct{} {
	list := bip39.GetWordList()
	set := make(map[string]struct{}, len(list))
	for _, w := range list {
		set[w] = struct{}{}
	}
	return set
}()

// ParseSecretRedactMode maps a config string onto a mode. Anything
// unrecognized — including the empty string — becomes hybrid, so a node that
// was never configured still scrubs.
func ParseSecretRedactMode(s string) SecretRedactMode {
	switch SecretRedactMode(strings.ToLower(strings.TrimSpace(s))) {
	case SecretRedactOff:
		return SecretRedactOff
	case SecretRedactStrict:
		return SecretRedactStrict
	default:
		return SecretRedactHybrid
	}
}

// RedactSecrets scrubs private keys and seed phrases from s, returning the
// cleaned string and how many secrets were replaced.
func RedactSecrets(s string, mode SecretRedactMode) (string, int) {
	if mode == SecretRedactOff || s == "" {
		return s, 0
	}
	// Seed phrases first: their spans are word runs, so scrubbing them cannot
	// create or destroy a hex match.
	out, seeds := redactSeedPhrases(s, mode)
	out, keys := redactHexSecrets(out, mode)
	return out, seeds + keys
}

// RedactMessagesSecrets scrubs every text-bearing field of msgs in place and
// returns the total number of secrets replaced. Tool-call arguments are
// included because an agent loop will happily round-trip a key through them.
func RedactMessagesSecrets(msgs []openai.ChatCompletionMessage, mode SecretRedactMode) int {
	if mode == SecretRedactOff {
		return 0
	}

	total := 0
	redact := func(field *string) {
		cleaned, n := RedactSecrets(*field, mode)
		if n > 0 {
			*field = cleaned
			total += n
		}
	}

	for i := range msgs {
		msg := &msgs[i]
		redact(&msg.Content)
		redact(&msg.ReasoningContent)
		for j := range msg.MultiContent {
			if msg.MultiContent[j].Type == openai.ChatMessagePartTypeText {
				redact(&msg.MultiContent[j].Text)
			}
		}
		if msg.FunctionCall != nil {
			redact(&msg.FunctionCall.Arguments)
		}
		for j := range msg.ToolCalls {
			redact(&msg.ToolCalls[j].Function.Arguments)
		}
	}
	return total
}

// redactHexSecrets replaces 64-hex blobs. In hybrid mode a 0x-prefixed blob is
// left alone unless a private-key cue sits within cueWindow bytes, on the
// theory that users paste keys bare and paste hashes with the prefix.
func redactHexSecrets(s string, mode SecretRedactMode) (string, int) {
	locs := hexSecretPattern.FindAllStringIndex(s, -1)
	if len(locs) == 0 {
		return s, 0
	}

	var b strings.Builder
	last, n := 0, 0
	for _, loc := range locs {
		start, end := loc[0], loc[1]
		if mode != SecretRedactStrict && hasHexPrefix(s[start:end]) && !hasPrivateKeyCue(s, start, end) {
			continue
		}
		b.WriteString(s[last:start])
		b.WriteString(RedactedPrivateKey)
		last = end
		n++
	}
	if n == 0 {
		return s, 0
	}
	b.WriteString(s[last:])
	return b.String(), n
}

func hasHexPrefix(match string) bool {
	return len(match) > 2 && match[0] == '0' && (match[1] == 'x' || match[1] == 'X')
}

// hasPrivateKeyCue reports whether a private-key cue appears within cueWindow
// bytes before or after the [start, end) span. Slicing on byte offsets can
// split a multi-byte rune at a window edge, which is harmless here: the cue
// pattern is pure ASCII and cannot match the resulting fragment.
func hasPrivateKeyCue(s string, start, end int) bool {
	from := start - cueWindow
	if from < 0 {
		from = 0
	}
	to := end + cueWindow
	if to > len(s) {
		to = len(s)
	}
	return privateKeyCuePattern.MatchString(s[from:start]) ||
		privateKeyCuePattern.MatchString(s[end:to])
}

// wordToken is one alphabetic run and where it sits in the source string.
type wordToken struct {
	word       string // lowercased, for wordlist lookup
	start, end int
}

// redactSeedPhrases replaces BIP-39 phrases. Candidates are maximal runs of
// adjacent wordlist words; within a run the scanner takes the longest legal
// phrase length that validates. Hybrid mode requires the BIP-39 checksum,
// which makes a false positive on ordinary prose effectively impossible;
// strict mode accepts any 12+ word run, catching phrases with a typo or a
// swapped word at the cost of occasionally eating a wordlist-heavy sentence.
func redactSeedPhrases(s string, mode SecretRedactMode) (string, int) {
	run := make([]wordToken, 0, 24)
	var spans [][2]int
	prevEnd := -1

	flush := func() {
		spans = append(spans, findPhraseSpans(run, mode)...)
		run = run[:0]
	}

	for _, loc := range wordPattern.FindAllStringIndex(s, -1) {
		start, end := loc[0], loc[1]
		word := strings.ToLower(s[start:end])
		_, known := bip39Words[word]
		adjacent := prevEnd >= 0 && start-prevEnd <= maxMnemonicGap

		if !known || (len(run) > 0 && !adjacent) {
			flush()
			if !known {
				prevEnd = end
				continue
			}
		}
		run = append(run, wordToken{word: word, start: start, end: end})
		prevEnd = end
	}
	flush()

	if len(spans) == 0 {
		return s, 0
	}

	var b strings.Builder
	last := 0
	for _, span := range spans {
		b.WriteString(s[last:span[0]])
		b.WriteString(RedactedSeedPhrase)
		last = span[1]
	}
	b.WriteString(s[last:])
	return b.String(), len(spans)
}

// findPhraseSpans walks one run of adjacent wordlist words and returns the
// byte spans of the phrases inside it.
func findPhraseSpans(run []wordToken, mode SecretRedactMode) [][2]int {
	var spans [][2]int
	for i := 0; i+minBip39Words <= len(run); {
		matched := 0
		for _, count := range bip39WordCounts {
			if i+count > len(run) {
				continue
			}
			if mode == SecretRedactStrict || bip39.IsMnemonicValid(joinWords(run[i:i+count])) {
				matched = count
				break
			}
		}
		if matched == 0 {
			i++
			continue
		}
		spans = append(spans, [2]int{run[i].start, run[i+matched-1].end})
		i += matched
	}
	return spans
}

func joinWords(tokens []wordToken) string {
	words := make([]string, len(tokens))
	for i, t := range tokens {
		words[i] = t.word
	}
	return strings.Join(words, " ")
}
