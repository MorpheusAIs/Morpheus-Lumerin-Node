package attestation

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

const rytnAttestationURL = "https://secretai-rytn.scrtlabs.com:29343"

func skipIfNoNetwork(t *testing.T) {
	if os.Getenv("TEE_INTEGRATION_TEST") == "" {
		t.Skip("set TEE_INTEGRATION_TEST=1 to run live integration tests")
	}
}

func insecureClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true, //nolint:gosec
			},
		},
	}
}

func fetchRaw(t *testing.T, url string) []byte {
	t.Helper()
	client := insecureClient()
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("GET %s returned %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading %s: %s", url, err)
	}
	return body
}

// TestIntegration_RytnLive fetches CPU quote, docker-compose, and registry
// from live endpoints and verifies the workload. Mirrors the app's AttestBackend flow.
func TestIntegration_RytnLive(t *testing.T) {
	skipIfNoNetwork(t)

	// 1. Fetch CPU quote from live endpoint
	cpuQuoteRaw := fetchRaw(t, rytnAttestationURL+"/cpu")
	cpuQuoteHex := strings.TrimSpace(string(cpuQuoteRaw))
	t.Logf("CPU quote: %d hex chars (%d bytes decoded)", len(cpuQuoteHex), len(cpuQuoteHex)/2)

	// 2. Parse TDX fields
	fields, err := ParseTdxQuoteFields(cpuQuoteHex)
	if err != nil {
		t.Fatalf("ParseTdxQuoteFields: %s", err)
	}
	t.Logf("MRTD:  %s", fields.MRTD)
	t.Logf("RTMR0: %s", fields.RTMR0)
	t.Logf("RTMR1: %s", fields.RTMR1)
	t.Logf("RTMR2: %s", fields.RTMR2)
	t.Logf("RTMR3: %s", fields.RTMR3)

	// 3. Fetch docker-compose from live endpoint (raw, no trimming)
	composeBody := fetchRaw(t, rytnAttestationURL+"/docker-compose")
	composeHash := sha256.Sum256(composeBody)
	t.Logf("docker-compose: %d bytes, sha256=%s", len(composeBody), hex.EncodeToString(composeHash[:]))
	t.Logf("docker-compose first 200 chars: %q", string(composeBody[:min(200, len(composeBody))]))
	t.Logf("docker-compose last 100 chars: %q", string(composeBody[max(0, len(composeBody)-100):]))

	// 4. Download registry from GitHub
	registryBody := fetchRaw(t, DefaultRegistryURL)
	entries := parseTdxRegistryCSV(string(registryBody))
	t.Logf("Registry: %d entries from %s", len(entries), DefaultRegistryURL)

	registry := NewArtifactRegistry("http://unused", 1*time.Hour, &lib.LoggerMock{})
	registry.mu.Lock()
	registry.entries = entries
	registry.lastFetched = time.Now()
	registry.mu.Unlock()

	// 5. Find matching artifacts
	candidates := registry.FindMatchingArtifacts(fields.MRTD, fields.RTMR0, fields.RTMR1, fields.RTMR2)
	t.Logf("Matching candidates: %d", len(candidates))

	for i, c := range candidates {
		calculated := CalculateRTMR3(composeBody, c.RootfsData)
		t.Logf("  candidate[%d]: template=%s ver=%s env=%s", i, c.TemplateName, c.ArtifactsVer, c.VMType)
		t.Logf("    rootfs_data:      %s", c.RootfsData)
		t.Logf("    calculated RTMR3: %s", calculated)
		t.Logf("    quote RTMR3:      %s", fields.RTMR3)
		t.Logf("    match:            %v", calculated == fields.RTMR3)
	}

	// 6. Run full VerifyWorkload
	result := VerifyWorkload(registry, cpuQuoteHex, string(composeBody), &lib.LoggerMock{})
	t.Logf("VerifyWorkload result: status=%s template=%s ver=%s env=%s", result.Status, result.TemplateName, result.ArtifactsVer, result.Env)

	if result.Status != WorkloadAuthentic {
		t.Logf("MISMATCH DETECTED - comparing with local file...")

		// Also try with the local file to see if it differs
		_, f, _, _ := runtime.Caller(0)
		localCompose, localErr := os.ReadFile(filepath.Join(filepath.Dir(f), "..", "..", "secretvm-verify", "rytn-docker-compose.yaml"))
		if localErr == nil {
			localHash := sha256.Sum256(localCompose)
			t.Logf("Local rytn-docker-compose.yaml: %d bytes, sha256=%s", len(localCompose), hex.EncodeToString(localHash[:]))
			localResult := VerifyWorkload(registry, cpuQuoteHex, string(localCompose), &lib.LoggerMock{})
			t.Logf("VerifyWorkload with local file: status=%s", localResult.Status)

			if len(composeBody) != len(localCompose) {
				t.Logf("SIZE DIFFERS: live=%d local=%d (diff=%d bytes)", len(composeBody), len(localCompose), len(composeBody)-len(localCompose))
			}
		}

		t.Fatalf("expected authentic_match, got %s", result.Status)
	}
}

// TestIntegration_RytnLocalFiles uses local cpu_quote.txt and rytn-docker-compose.yaml
// with the live registry from GitHub.
func TestIntegration_RytnLocalFiles(t *testing.T) {
	_, f, _, _ := runtime.Caller(0)
	base := filepath.Join(filepath.Dir(f), "..", "..", "secretvm-verify")

	quoteBytes, err := os.ReadFile(filepath.Join(base, "cpu_quote.txt"))
	if err != nil {
		t.Skipf("cpu_quote.txt not found: %s", err)
	}
	composeBytes, err := os.ReadFile(filepath.Join(base, "rytn-docker-compose.yaml"))
	if err != nil {
		t.Skipf("rytn-docker-compose.yaml not found: %s", err)
	}

	cpuQuoteHex := strings.TrimSpace(string(quoteBytes))

	fields, err := ParseTdxQuoteFields(cpuQuoteHex)
	if err != nil {
		t.Fatalf("ParseTdxQuoteFields: %s", err)
	}

	composeHash := sha256.Sum256(composeBytes)
	t.Logf("Local compose: %d bytes, sha256=%s", len(composeBytes), hex.EncodeToString(composeHash[:]))
	t.Logf("Quote RTMR3: %s", fields.RTMR3)

	// Try with local CSV first
	csvBytes, csvErr := os.ReadFile(filepath.Join(base, "artifacts_registry", "tdx.csv"))
	if csvErr != nil {
		t.Skipf("tdx.csv not found: %s", csvErr)
	}

	registry := NewArtifactRegistry("http://unused", 1*time.Hour, &lib.LoggerMock{})
	entries := parseTdxRegistryCSV(string(csvBytes))
	registry.mu.Lock()
	registry.entries = entries
	registry.lastFetched = time.Now()
	registry.mu.Unlock()
	t.Logf("Local registry: %d entries", len(entries))

	candidates := registry.FindMatchingArtifacts(fields.MRTD, fields.RTMR0, fields.RTMR1, fields.RTMR2)
	t.Logf("Matching candidates: %d", len(candidates))

	for i, c := range candidates {
		calculated := CalculateRTMR3(composeBytes, c.RootfsData)
		t.Logf("  candidate[%d]: template=%s ver=%s rootfs=%s", i, c.TemplateName, c.ArtifactsVer, c.RootfsData)
		t.Logf("    calculated=%s quote=%s match=%v", calculated, fields.RTMR3, calculated == fields.RTMR3)
	}

	result := VerifyWorkload(registry, cpuQuoteHex, string(composeBytes), &lib.LoggerMock{})
	t.Logf("Result: status=%s template=%s ver=%s env=%s", result.Status, result.TemplateName, result.ArtifactsVer, result.Env)

	if result.Status != WorkloadAuthentic {
		t.Fatalf("expected authentic_match, got %s", result.Status)
	}
}

// TestIntegration_RytnBackendVerifier runs the full BackendVerifier.AttestBackend flow
// against the live rytn endpoint, mirroring exactly what the app does.
func TestIntegration_RytnBackendVerifier(t *testing.T) {
	skipIfNoNetwork(t)

	ctx := context.Background()

	registry := NewArtifactRegistry(DefaultRegistryURL, 1*time.Hour, &lib.LoggerMock{})
	registry.Start(ctx)

	if !registry.IsLoaded() {
		t.Fatal("registry failed to load")
	}
	t.Logf("Registry loaded with entries")

	bv := NewBackendVerifier(
		DefaultPortalURL,
		&NoopGoldenSource{},
		registry,
		&lib.LoggerMock{},
	)

	err := bv.AttestBackend(ctx, "test-model", rytnAttestationURL)
	if err != nil {
		t.Logf("AttestBackend error: %s", err)

		status := bv.GetStatus("test-model")
		if status != nil {
			t.Logf("Status: %s, Error: %s", status.Status, status.Error)
			t.Logf("WorkloadStatus: %s, Template: %s, Ver: %s", status.WorkloadStatus, status.VMTemplateName, status.ArtifactsVersion)
		}

		t.Fatalf("AttestBackend failed: %s", err)
	}

	status := bv.GetStatus("test-model")
	t.Logf("SUCCESS: status=%s workload=%s template=%s ver=%s",
		status.Status, status.WorkloadStatus, status.VMTemplateName, status.ArtifactsVersion)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func init() {
	// Suppress unused import error for fmt
	_ = fmt.Sprintf
}
