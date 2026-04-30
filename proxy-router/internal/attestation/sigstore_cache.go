package attestation

import (
	"path/filepath"
	"sync"

	"github.com/sigstore/sigstore-go/pkg/tuf"
)

// sigstoreCacheDir, when non-empty, overrides the default TUF cache location
// (`$HOME/.sigstore/root`) used by [sigstore-go] when fetching the trusted
// root. The default path blows up inside the iOS app sandbox — the home
// directory returned by `os.UserHomeDir()` maps to
// `/private/var/mobile/Containers/Data/Application/<UUID>/` which is
// read-only at the top level; only `Documents/` and `Library/` are
// writable. Mobile embedders call [SetSigstoreCacheDir] during SDK init to
// point TUF at the app's writable data directory instead.
var (
	sigstoreCacheMu  sync.RWMutex
	sigstoreCacheDir string
)

// SetSigstoreCacheDir configures a writable directory to be used as the
// Sigstore TUF metadata cache. Pass an empty string to fall back to the
// library default ($HOME/.sigstore/root), which is the right choice on
// desktop and server builds.
//
// Safe to call from any goroutine; takes effect on the next trusted-root
// fetch. The supplied path is joined with a `sigstore` subdirectory so
// mobile apps can share a single data-dir across multiple Go caches
// without them colliding.
func SetSigstoreCacheDir(dir string) {
	sigstoreCacheMu.Lock()
	defer sigstoreCacheMu.Unlock()
	if dir == "" {
		sigstoreCacheDir = ""
		return
	}
	sigstoreCacheDir = filepath.Join(dir, "sigstore")
}

// tufOptions returns a fresh [tuf.Options] with the sandbox-safe cache path
// applied if one has been configured. Callers should treat the result as
// ephemeral and not cache it between fetches — the underlying options hold
// a mutable CachePath.
func tufOptions() *tuf.Options {
	opts := tuf.DefaultOptions()
	sigstoreCacheMu.RLock()
	dir := sigstoreCacheDir
	sigstoreCacheMu.RUnlock()
	if dir != "" {
		opts = opts.WithCachePath(dir)
	}
	return opts
}
