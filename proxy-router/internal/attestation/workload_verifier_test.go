package attestation

import (
	"encoding/hex"
	"html"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

const (
	testDataRelPath    = "../../secretvm-verify/test-data"
	registryCSVRelPath = "../../secretvm-verify/artifacts_registry/tdx.csv"
)

func readTestFixture(t *testing.T, filename string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(testDataRelPath, filename))
	if err != nil {
		t.Fatalf("read fixture %s: %v", filename, err)
	}
	return strings.TrimSpace(string(data))
}

func testRegistry(t *testing.T) *ArtifactRegistry {
	t.Helper()
	csvData, err := os.ReadFile(registryCSVRelPath)
	if err != nil {
		t.Fatalf("read registry CSV: %v", err)
	}
	reg := NewArtifactRegistry("", 0, &lib.LoggerMock{})
	reg.entries = parseTdxRegistryCSV(string(csvData))
	reg.lastFetched = time.Now()
	return reg
}

func TestParseTdxQuoteFields_Valid(t *testing.T) {
	quoteHex := readTestFixture(t, "tdx_cpu_docker_check_quote.txt")
	fields, err := ParseTdxQuoteFields(quoteHex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, tc := range []struct {
		name string
		val  string
	}{
		{"MRTD", fields.MRTD},
		{"RTMR0", fields.RTMR0},
		{"RTMR1", fields.RTMR1},
		{"RTMR2", fields.RTMR2},
	} {
		if len(tc.val) != 96 {
			t.Fatalf("%s length = %d, want 96", tc.name, len(tc.val))
		}
		if _, err := hex.DecodeString(tc.val); err != nil {
			t.Fatalf("%s not valid hex: %v", tc.name, err)
		}
	}
	if fields.RTMR3 == "" {
		t.Fatal("RTMR3 is empty")
	}
}

func TestParseTdxQuoteFields_TooShort(t *testing.T) {
	_, err := ParseTdxQuoteFields("0400020081000000")
	if err == nil {
		t.Fatal("expected error for short quote")
	}
}

func TestParseTdxQuoteFields_InvalidHex(t *testing.T) {
	_, err := ParseTdxQuoteFields("zzzz_not_hex_at_all!!!")
	if err == nil {
		t.Fatal("expected error for invalid hex")
	}
}

func TestIsTdxQuote_Valid(t *testing.T) {
	quoteHex := readTestFixture(t, "tdx_cpu_docker_check_quote.txt")
	if !IsTdxQuote(quoteHex) {
		t.Fatal("expected true for valid TDX quote")
	}
}

func TestIsTdxQuote_Invalid(t *testing.T) {
	if IsTdxQuote("not_a_quote_at_all") {
		t.Fatal("expected false for garbage input")
	}
}

func TestCalculateRTMR3_Deterministic(t *testing.T) {
	compose := []byte("services:\n  app:\n    image: test:latest\n")
	rootfs := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	r1 := CalculateRTMR3(compose, rootfs)
	r2 := CalculateRTMR3(compose, rootfs)
	if r1 != r2 {
		t.Fatalf("non-deterministic: %s != %s", r1, r2)
	}
	if len(r1) != 96 {
		t.Fatalf("result length = %d, want 96", len(r1))
	}
}

func TestReplayRtmr_EmptyLog(t *testing.T) {
	result := replayRtmr(nil)
	want := strings.Repeat("0", 96)
	if result != want {
		t.Fatalf("got %s, want %s", result, want)
	}
}

func TestVerifyTdxWorkload_AuthenticMatch(t *testing.T) {
	reg := testRegistry(t)
	quoteHex := readTestFixture(t, "tdx_cpu_docker_check_quote.txt")
	composeYaml := readTestFixture(t, "tdx_cpu_docker_check_compose.yaml")

	result := VerifyTdxWorkload(reg, quoteHex, composeYaml, nil)
	if result.Status != WorkloadAuthentic {
		t.Fatalf("status = %s, want %s", result.Status, WorkloadAuthentic)
	}
}

// TestVerifyTdxWorkload_HTMLWrappedCompose verifies that a docker-compose
// served by old attest-rest (HTML page with the YAML in a <pre> block) still
// matches: the HTML-extracted variant is tried alongside the raw response.
func TestVerifyTdxWorkload_HTMLWrappedCompose(t *testing.T) {
	reg := testRegistry(t)
	quoteHex := readTestFixture(t, "tdx_cpu_docker_check_quote.txt")
	composeYaml := readTestFixture(t, "tdx_cpu_docker_check_compose.yaml")
	wrapped := "<html><body><pre>" + html.EscapeString(composeYaml) + "</pre></body></html>"

	result := VerifyTdxWorkload(reg, quoteHex, wrapped, nil)
	if result.Status != WorkloadAuthentic {
		t.Fatalf("status = %s, want %s", result.Status, WorkloadAuthentic)
	}
}

func TestComposeCandidates_Raw(t *testing.T) {
	raw := "services:\n  app:\n    image: test\n"
	got := composeCandidates(raw)
	if len(got) != 1 || got[0] != raw {
		t.Fatalf("expected single raw candidate, got %q", got)
	}
}

func TestComposeCandidates_HTMLWrapped(t *testing.T) {
	inner := "services:\n  app:\n    image: test\n"
	wrapped := "<pre>" + html.EscapeString(inner) + "</pre>"
	got := composeCandidates(wrapped)
	if len(got) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(got))
	}
	if got[0] != wrapped {
		t.Fatalf("first candidate must be the raw response, got %q", got[0])
	}
	if got[1] != inner {
		t.Fatalf("second candidate must be the extracted content, got %q", got[1])
	}
}

func TestVerifyTdxWorkload_AuthenticMismatch(t *testing.T) {
	reg := testRegistry(t)
	quoteHex := readTestFixture(t, "tdx_cpu_docker_check_quote.txt")
	composeYaml := readTestFixture(t, "tdx_cpu_docker_check_compose.yaml") + "\n# tampered"

	result := VerifyTdxWorkload(reg, quoteHex, composeYaml, nil)
	if result.Status != WorkloadAuthenticMismatch {
		t.Fatalf("status = %s, want %s", result.Status, WorkloadAuthenticMismatch)
	}
}

func TestVerifyTdxWorkload_NotAuthentic(t *testing.T) {
	reg := testRegistry(t)
	quoteHex := readTestFixture(t, "tdx_cpu_docker_check_quote.txt")

	raw, err := hex.DecodeString(quoteHex)
	if err != nil {
		t.Fatalf("decode quote: %v", err)
	}
	raw[184] ^= 0xFF
	raw[185] ^= 0xFF
	corrupted := hex.EncodeToString(raw)

	result := VerifyTdxWorkload(reg, corrupted, "anything", nil)
	if result.Status != WorkloadNotAuthentic {
		t.Fatalf("status = %s, want %s", result.Status, WorkloadNotAuthentic)
	}
}

func TestVerifyWorkload_TdxQuote(t *testing.T) {
	reg := testRegistry(t)
	quoteHex := readTestFixture(t, "tdx_cpu_docker_check_quote.txt")
	composeYaml := readTestFixture(t, "tdx_cpu_docker_check_compose.yaml")

	result := VerifyWorkload(reg, nil, quoteHex, composeYaml, nil)
	if result.Status != WorkloadAuthentic {
		t.Fatalf("status = %s, want %s", result.Status, WorkloadAuthentic)
	}
}

func TestVerifyWorkload_TdxQuoteRegistryNil(t *testing.T) {
	quoteHex := readTestFixture(t, "tdx_cpu_docker_check_quote.txt")
	composeYaml := readTestFixture(t, "tdx_cpu_docker_check_compose.yaml")

	// TDX quote but no artifact registry -> must fail closed.
	result := VerifyWorkload(nil, nil, quoteHex, composeYaml, nil)
	if result.Status != ArtifactRegistryNotAvailable {
		t.Fatalf("status = %s, want %s", result.Status, ArtifactRegistryNotAvailable)
	}
}

func TestVerifyWorkload_TdxQuoteRegistryNotLoaded(t *testing.T) {
	quoteHex := readTestFixture(t, "tdx_cpu_docker_check_quote.txt")
	composeYaml := readTestFixture(t, "tdx_cpu_docker_check_compose.yaml")

	// Registry constructed but never fetched -> IsLoaded()==false -> fail closed.
	unloaded := NewArtifactRegistry("", 0, &lib.LoggerMock{})
	result := VerifyWorkload(unloaded, nil, quoteHex, composeYaml, nil)
	if result.Status != ArtifactRegistryNotAvailable {
		t.Fatalf("status = %s, want %s", result.Status, ArtifactRegistryNotAvailable)
	}
}

func TestVerifyWorkload_NonTdxQuoteNoSevRegistry(t *testing.T) {
	reg := testRegistry(t)
	// SEV quote with no SEV registry available must fail closed (registry unavailable).
	result := VerifyWorkload(reg, nil, "SGVsbG8gV29ybGQ=", "anything", nil)
	if result.Status != ArtifactRegistryNotAvailable {
		t.Fatalf("status = %s, want %s", result.Status, ArtifactRegistryNotAvailable)
	}
}
