package proxyapi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateAllowlistedFilePath(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "a", "b.bin")
	require.NoError(t, os.MkdirAll(filepath.Dir(inside), 0o700))
	require.NoError(t, os.WriteFile(inside, []byte("x"), 0o600))

	require.NoError(t, ValidateAllowlistedFilePath(inside, []string{root}))

	outside := filepath.Join(t.TempDir(), "other.bin")
	require.NoError(t, os.WriteFile(outside, []byte("y"), 0o600))
	require.Error(t, ValidateAllowlistedFilePath(outside, []string{root}))

	dd := string([]byte{46, 46})
	escaped := filepath.Join(root, dd, "etc", "x")
	require.Error(t, ValidateAllowlistedFilePath(escaped, []string{root}))
	require.Error(t, ValidateAllowlistedFilePath("", []string{root}))
}
