package proxyapi

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveIPFSDownloadDest_RejectsMalformed(t *testing.T) {
	root := t.TempDir()
	sep := string(filepath.Separator)
	// Spelled by hand: filepath.Join would clean "." and ".." away.
	for _, dest := range []string{"", "   ", "relative/file", "./x", sep, root + sep, root + sep + ".", root + sep + ".."} {
		_, err := resolveIPFSDownloadDest(dest, true, nil)
		assert.ErrorIs(t, err, errIPFSPathInvalid, "dest=%q", dest)
	}
}

func TestResolveIPFSDownloadDest_FullAccessMayWriteAnywhere(t *testing.T) {
	root := t.TempDir()
	got, err := resolveIPFSDownloadDest(filepath.Join(root, "a", "..", "b", "model.gguf"), true, nil)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, "b", "model.gguf"), got)
}

func TestResolveIPFSDownloadDest_ScopedCallerConfinedToAllowedDirs(t *testing.T) {
	allowed := t.TempDir()
	outside := t.TempDir()

	_, err := resolveIPFSDownloadDest(filepath.Join(allowed, "m.gguf"), false, nil)
	assert.ErrorIs(t, err, errIPFSPathNotAllowed, "no dirs configured")

	got, err := resolveIPFSDownloadDest(filepath.Join(allowed, "m.gguf"), false, []string{allowed})
	require.NoError(t, err)
	realAllowed, err := filepath.EvalSymlinks(allowed)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(realAllowed, "m.gguf"), got)

	// The operator may configure the resolved spelling, the caller may use the
	// symlinked one (macOS: /var vs /private/var), and the reverse.
	_, err = resolveIPFSDownloadDest(filepath.Join(allowed, "m.gguf"), false, []string{realAllowed})
	assert.NoError(t, err)
	_, err = resolveIPFSDownloadDest(filepath.Join(realAllowed, "m.gguf"), false, []string{allowed})
	assert.NoError(t, err)

	_, err = resolveIPFSDownloadDest(filepath.Join(outside, "m.gguf"), false, []string{allowed})
	assert.ErrorIs(t, err, errIPFSPathNotAllowed)

	_, err = resolveIPFSDownloadDest(filepath.Join(allowed, "..", filepath.Base(outside), "m.gguf"), false, []string{allowed})
	assert.ErrorIs(t, err, errIPFSPathNotAllowed, "dot-dot escape")

	_, err = resolveIPFSDownloadDest(filepath.Join(allowed, "missing", "m.gguf"), false, []string{allowed})
	assert.ErrorIs(t, err, errIPFSPathNotAllowed, "parent directory must exist")
}

func TestResolveIPFSDownloadDest_SymlinkInsideAllowedDirCannotEscape(t *testing.T) {
	allowed := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.Symlink(outside, filepath.Join(allowed, "link")))

	_, err := resolveIPFSDownloadDest(filepath.Join(allowed, "link", "m.gguf"), false, []string{allowed})
	assert.ErrorIs(t, err, errIPFSPathNotAllowed)
}

func TestResolveIPFSSourcePath_ScopedCallerConfinedToAllowedDirs(t *testing.T) {
	allowed := t.TempDir()
	outside := t.TempDir()
	inside := filepath.Join(allowed, "model.bin")
	secret := filepath.Join(outside, ".cookie")
	require.NoError(t, os.WriteFile(inside, []byte("x"), 0o600))
	require.NoError(t, os.WriteFile(secret, []byte("admin:pw"), 0o600))
	require.NoError(t, os.Symlink(secret, filepath.Join(allowed, "innocent.bin")))

	got, err := resolveIPFSSourcePath(inside, false, []string{allowed})
	require.NoError(t, err)
	realInside, _ := filepath.EvalSymlinks(inside)
	assert.Equal(t, realInside, got)

	_, err = resolveIPFSSourcePath(secret, false, []string{allowed})
	assert.ErrorIs(t, err, errIPFSPathNotAllowed)

	_, err = resolveIPFSSourcePath(filepath.Join(allowed, "innocent.bin"), false, []string{allowed})
	assert.ErrorIs(t, err, errIPFSPathNotAllowed, "symlink to a file outside")

	_, err = resolveIPFSSourcePath(filepath.Join(allowed, "nope.bin"), false, []string{allowed})
	assert.ErrorIs(t, err, errIPFSPathNotAllowed, "file must exist")

	got, err = resolveIPFSSourcePath(secret, true, nil)
	require.NoError(t, err)
	assert.Equal(t, secret, got, "full access is not confined")
}

func TestSetIPFSAllowedDirs_SplitsPathList(t *testing.T) {
	c := &ProxyController{}
	c.SetIPFSAllowedDirs(strings.Join([]string{"/a", " ", "/b"}, string(filepath.ListSeparator)))
	assert.Equal(t, []string{"/a", "/b"}, c.ipfsAllowedDirs)
	c.SetIPFSAllowedDirs("")
	assert.Empty(t, c.ipfsAllowedDirs)
}

func TestCreateDownloadFile_NeverOverwrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "m.gguf")
	f, err := createDownloadFile(path)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	_, err = createDownloadFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// The path checks run before IPFS is touched, so these pass with IPFS disabled.
func TestIPFSHandlers_AuthorizePathsBeforeIPFS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	cookiePath := filepath.Join(dir, ".cookie")
	authCfg := system.NewAuthConfig(filepath.Join(dir, "proxy.conf"), cookiePath, "admin:admin-pw", nil)
	require.NoError(t, authCfg.EnsureConfigFilesExist())
	require.NoError(t, authCfg.AddUser("agent", "agent-pw", []string{"ipfs_get", "ipfs_add"}))

	log := lib.NewTestLogger()
	ctrl := NewProxyController(nil, nil, nil, false, false, *authCfg, NewIpfsManagerDisabled(log), log)
	allowed := t.TempDir()
	ctrl.SetIPFSAllowedDirs(allowed)
	insideFile := filepath.Join(allowed, "model.bin")
	require.NoError(t, os.WriteFile(insideFile, []byte("x"), 0o600))

	r := gin.New()
	r.GET("/ipfs/download/:cidHash", authCfg.CheckAuth("ipfs_get"), ctrl.DownloadFile)
	r.GET("/ipfs/download/stream/:cidHash", authCfg.CheckAuth("ipfs_get"), ctrl.StreamDownloadFile)
	r.POST("/ipfs/add", authCfg.CheckAuth("ipfs_add"), ctrl.AddFile)

	hash := "0x" + strings.Repeat("ab", 32)
	do := func(method, path, user, pass, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(user+":"+pass)))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	for _, route := range []string{"/ipfs/download/", "/ipfs/download/stream/"} {
		w := do(http.MethodGet, route+hash+"?dest="+url.QueryEscape(filepath.Join(dir, "evil")), "agent", "agent-pw", "")
		assert.Equal(t, http.StatusForbidden, w.Code, route+" scoped user outside allowed dirs: "+w.Body.String())

		w = do(http.MethodGet, route+hash+"?dest=relative.bin", "admin", "admin-pw", "")
		assert.Equal(t, http.StatusBadRequest, w.Code, route+" relative dest")
		assert.Contains(t, w.Body.String(), "absolute")
	}

	// Inside the allowed dir the scoped user passes validation and reaches the
	// (disabled) IPFS manager; full access is never confined.
	w := do(http.MethodGet, "/ipfs/download/"+hash+"?dest="+url.QueryEscape(filepath.Join(allowed, "ok.bin")), "agent", "agent-pw", "")
	assert.Contains(t, w.Body.String(), "IPFS node is not ready")
	w = do(http.MethodGet, "/ipfs/download/"+hash+"?dest="+url.QueryEscape(filepath.Join(dir, "anywhere.bin")), "admin", "admin-pw", "")
	assert.Contains(t, w.Body.String(), "IPFS node is not ready")

	// ipfs_add: a scoped user cannot publish the cookie file.
	addBody := func(path string) string {
		return fmt.Sprintf(`{"filePath":%q,"tags":[],"id":%q,"modelName":"m"}`, path, hash)
	}
	w = do(http.MethodPost, "/ipfs/add", "agent", "agent-pw", addBody(cookiePath))
	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	w = do(http.MethodPost, "/ipfs/add", "agent", "agent-pw", addBody(insideFile))
	assert.Contains(t, w.Body.String(), "IPFS node is not ready")
	w = do(http.MethodPost, "/ipfs/add", "admin", "admin-pw", addBody(cookiePath))
	assert.Contains(t, w.Body.String(), "IPFS node is not ready")
}
