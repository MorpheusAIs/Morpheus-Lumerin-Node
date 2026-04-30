package mobile

import "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/attestation"

// SetSigstoreCacheDir re-exports [attestation.SetSigstoreCacheDir] so
// mobile embedders can configure the Sigstore TUF cache location without
// reaching into the proxy-router `internal/` tree (which is disallowed by
// Go's internal package rule).
//
// Pass the app's writable data directory on init; sigstore-go's default of
// `$HOME/.sigstore/root` is unusable on iOS because the home directory
// returned by `os.UserHomeDir()` is the sandboxed app container root and
// is read-only at the top level.
func SetSigstoreCacheDir(dir string) {
	attestation.SetSigstoreCacheDir(dir)
}
