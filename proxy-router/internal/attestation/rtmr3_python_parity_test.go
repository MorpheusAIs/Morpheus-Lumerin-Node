package attestation

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComputeRTMR3PythonParity runs proxy-router/scripts/compute-rtmr3.py
// (the CI/CD tool that bakes the TDX golden into the signed manifest) against
// the SCRT ground-truth dstack vectors and the attest-tool scheme, and asserts
// it agrees byte-for-byte with the runtime Go implementation in rtmr.go.
//
// Skipped when python3 is not on PATH.
func TestComputeRTMR3PythonParity(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not on PATH; skipping parity test")
	}

	_, thisFile, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	scriptPath := filepath.Join(repoRoot, "proxy-router", "scripts", "compute-rtmr3.py")
	_, err := os.Stat(scriptPath)
	require.NoError(t, err, "compute-rtmr3.py not found")

	composePath := filepath.Join(t.TempDir(), "docker-compose.yaml")
	require.NoError(t, os.WriteFile(composePath, []byte(dstackFixtureCompose), 0o600))

	type out struct {
		RTMR3           string `json:"rtmr3"`
		Scheme          string `json:"rtmr3_scheme"`
		ComposeSHA256   string `json:"compose_sha256"`
		RootfsSHA256    string `json:"rootfs_sha256"`
		DstackAppID     string `json:"dstack_app_id"`
		RTMR3AttestTool string `json:"rtmr3_attest_tool"`
	}
	run := func(args ...string) out {
		cmd := exec.Command("python3", append([]string{scriptPath, composePath}, args...)...)
		raw, err := cmd.CombinedOutput()
		require.NoError(t, err, "script failed: %s", raw)
		var o out
		require.NoError(t, json.Unmarshal(raw, &o), "bad JSON: %s", raw)
		return o
	}

	for _, v := range dstackVectors {
		t.Run("dstack/"+v.name, func(t *testing.T) {
			o := run("--rootfs-sha256="+v.rootfs, "--dstack-app-id=0x"+v.appID, "--json")
			assert.Equal(t, v.rtmr3, o.RTMR3, "python dstack replay must match the live quote")
			assert.Equal(t, "dstack", o.Scheme)
			assert.Equal(t, dstackFixtureComposeSHA256, o.ComposeSHA256)
			assert.Equal(t, v.appID, o.DstackAppID)
			assert.Equal(t, CalculateRTMR3FromHashes(dstackFixtureComposeSHA256, v.rootfs, ""), o.RTMR3AttestTool)
		})
	}

	t.Run("attest-tool", func(t *testing.T) {
		rootfs := dstackVectors[0].rootfs
		o := run("--rootfs-sha256="+rootfs, "--json")
		assert.Equal(t, CalculateRTMR3([]byte(dstackFixtureCompose), rootfs), o.RTMR3)
		assert.Equal(t, "attest-tool", o.Scheme)
		assert.Equal(t, "", o.DstackAppID)
	})
}
