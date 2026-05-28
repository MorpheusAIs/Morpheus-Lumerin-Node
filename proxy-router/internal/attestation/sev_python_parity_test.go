package attestation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// parityFixtureArtifactsVer is the synthetic registry version used by the
// parity test. It is intentionally NOT a real SCRT Labs release — the parity
// test exercises algorithmic equivalence between sev_gctx.go and
// compute-sev-measurement.py, NOT correctness of any specific upstream
// release. Future SecretVM bumps in `.github/tee/secretvm.env` should NOT
// require touching this file.
const parityFixtureArtifactsVer = "parity-test-fixture-v1"

// parityFixtureSevEntry is a frozen, hermetic SEV registry entry used to
// drive the Go ↔ Python parity test. The byte values are plausible-shaped
// (correct lengths and a realistic OVMF section table) but are NOT pinned
// to any particular SecretVM release. The parity assertion checks only that
// both implementations produce identical SHA-384 chains given identical
// inputs — so any well-formed entry suffices.
//
// Sourced from scrtlabs/secretvm-verify SEV registry as of 2026-04-29 and
// then frozen here. Keep this fixture stable; if you need to test against
// the live production registry, write a separate (network-gated)
// integration test.
var parityFixtureSevEntry = &SevArtifactEntry{
	VMType:            "prod",
	ArtifactsVer:      parityFixtureArtifactsVer,
	KernelHash:        "98c41a86a1ba6a9a9d772ae0b028835091b4930f79ea509b595d2080d7df90c2",
	InitrdHash:        "6b19d1b356c1e791f5c1c3d7dd86a723b870c87ebdfe3f7ccace80215ac71d2e",
	VcpuType:          "EPYC",
	RootfsHash:        "f44141c9a0cbed19ddf30b16929de33633e5631cfd68731e9ff9e4321d5775fd",
	OvmfHash:          "c581d3eaebf2941beb1f757de97497279b953a6999921cab05f9ed5268f9c0505d741f4021b5a3995c9893851cde190e",
	SevHashesTableGPA: 8457216,
	SevEsResetEIP:     8433668,
	OvmfSections: []SevOvmfSection{
		{GPA: 8388608, Size: 36864, SectionType: 1},
		{GPA: 8429568, Size: 12288, SectionType: 1},
		{GPA: 8441856, Size: 4096, SectionType: 2},
		{GPA: 8445952, Size: 4096, SectionType: 3},
		{GPA: 8450048, Size: 4096, SectionType: 4},
		{GPA: 8458240, Size: 61440, SectionType: 1},
		{GPA: 8454144, Size: 4096, SectionType: 16},
	},
}

// TestComputeSevMeasurementPythonParity runs proxy-router/scripts/compute-sev-measurement.py
// against a hermetic fixture and confirms it produces the same per-template launch
// digests as the Go implementation in sev_gctx.go (CalcSevMeasurement). This is the
// regression guard that keeps the CI/CD pipeline's Python tool in lockstep with the
// runtime Go source-of-truth.
//
// The fixture (parityFixtureSevEntry) is intentionally version-agnostic — its
// `artifacts_ver` is "parity-test-fixture-v1", NOT a real SecretVM release. Bumping
// the production pin in `.github/tee/secretvm.env` does NOT require touching this
// test; the test only proves "given the same inputs, both implementations produce
// the same SHA-384 chains". For a freshness check against the live SCRT Labs
// registry, write a separate (network-gated) integration test.
//
// Skipped when python3 is not on PATH (e.g. local Windows dev box).
func TestComputeSevMeasurementPythonParity(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not on PATH; skipping parity test")
	}

	_, thisFile, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..") // .../Morpheus-Lumerin-Node
	scriptPath := filepath.Join(repoRoot, "proxy-router", "scripts", "compute-sev-measurement.py")
	if _, err := os.Stat(scriptPath); err != nil {
		t.Fatalf("compute-sev-measurement.py not found at %s: %s", scriptPath, err)
	}

	tmp := t.TempDir()
	registryPath := filepath.Join(tmp, "sev.json")
	registryJSON, err := json.Marshal([]*SevArtifactEntry{parityFixtureSevEntry})
	if err != nil {
		t.Fatalf("marshal registry fixture: %s", err)
	}
	if err := os.WriteFile(registryPath, registryJSON, 0o600); err != nil {
		t.Fatalf("write registry fixture: %s", err)
	}

	composeContent := []byte("services:\n  proxy-router:\n    image: ghcr.io/example/test:latest\n")
	composePath := filepath.Join(tmp, "docker-compose.tee.yml")
	if err := os.WriteFile(composePath, composeContent, 0o600); err != nil {
		t.Fatalf("write compose fixture: %s", err)
	}

	// Compute the Go-side expected values against the *actual* compose bytes.
	composeHash := sha256.Sum256(composeContent)
	composeShaHex := hex.EncodeToString(composeHash[:])
	cmdline := fmt.Sprintf(
		"console=ttyS0 loglevel=7 docker_compose_hash=%s rootfs_hash=%s",
		composeShaHex, parityFixtureSevEntry.RootfsHash,
	)

	templates := []struct {
		name  string
		vcpus int
	}{
		{"small", 1},
		{"medium", 2},
		{"large", 4},
		{"2xlarge", 8},
		{"4xlarge", 16},
	}

	goExpected := make(map[string]string, len(templates))
	for _, tmpl := range templates {
		goExpected[tmpl.name] = CalcSevMeasurement(parityFixtureSevEntry, tmpl.vcpus, cmdline)
		if len(goExpected[tmpl.name]) != 96 {
			t.Fatalf("Go CalcSevMeasurement returned non-SHA-384 length for %s: %d hex chars",
				tmpl.name, len(goExpected[tmpl.name]))
		}
	}

	cmd := exec.Command("python3", scriptPath,
		"--registry", registryPath,
		"--vm-type", parityFixtureSevEntry.VMType,
		"--artifacts-ver", parityFixtureArtifactsVer,
		"--compose", composePath,
		"--all-templates", "--json",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python script failed: %s\n--- output ---\n%s", err, out)
	}

	var py struct {
		ComposeSHA256 string            `json:"compose_sha256"`
		PerTemplate   map[string]string `json:"per_template"`
	}
	if err := json.Unmarshal(out, &py); err != nil {
		t.Fatalf("decode python output: %s\n--- output ---\n%s", err, out)
	}

	if !strings.EqualFold(py.ComposeSHA256, composeShaHex) {
		t.Fatalf("compose_sha256 mismatch: python=%s go=%s", py.ComposeSHA256, composeShaHex)
	}

	for _, tmpl := range templates {
		got := py.PerTemplate[tmpl.name]
		want := goExpected[tmpl.name]
		if !strings.EqualFold(got, want) {
			t.Fatalf("SEV measurement mismatch for %s (vcpus=%d):\n  go     = %s\n  python = %s",
				tmpl.name, tmpl.vcpus, want, got)
		}
	}
}
