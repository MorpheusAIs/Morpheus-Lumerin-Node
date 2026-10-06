package proxyapi

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

var (
	// errIPFSPathInvalid: the path is not an absolute file path.
	errIPFSPathInvalid = errors.New("invalid file path")
	// errIPFSPathNotAllowed: a caller without full API access named a path
	// outside the configured IPFS directories.
	errIPFSPathNotAllowed = errors.New("file path not allowed")
)

// SetIPFSAllowedDirs configures the directories that callers without full API
// access may read from (ipfs_add) and write into (ipfs_get), given as an OS
// path list (IPFS_ALLOWED_DIRS). Blank entries are dropped.
func (c *ProxyController) SetIPFSAllowedDirs(pathList string) {
	c.ipfsAllowedDirs = nil
	for _, dir := range filepath.SplitList(pathList) {
		if dir = strings.TrimSpace(dir); dir != "" {
			c.ipfsAllowedDirs = append(c.ipfsAllowedDirs, dir)
		}
	}
}

// resolveIPFSDownloadDest validates the `dest` of an IPFS download and
// returns the path to create. The file does not exist yet, so only its
// directory is resolved through symlinks. Overwrite protection lives in
// createDownloadFile (O_EXCL).
func resolveIPFSDownloadDest(dest string, fullAccess bool, allowedDirs []string) (string, error) {
	return resolveIPFSPath(dest, fullAccess, allowedDirs, resolveParentDir)
}

// resolveIPFSSourcePath validates the `filePath` of an IPFS add. Without it
// ipfs_add let any holder publish .cookie (the admin password) or proxy.conf
// to IPFS. The file must exist, so it is resolved through symlinks itself.
func resolveIPFSSourcePath(path string, fullAccess bool, allowedDirs []string) (string, error) {
	return resolveIPFSPath(path, fullAccess, allowedDirs, filepath.EvalSymlinks)
}

// resolveIPFSPath applies the shared rules: the path must be absolute and name
// a file. A caller with full API access (the operator, or the desktop app
// acting for them) may use any such path, which keeps the desktop's "pick a
// folder" flow working. Any other caller is confined to allowedDirs and
// refused when none are configured, so ipfs_get and ipfs_add can be delegated
// to an agent without handing it the node's filesystem.
func resolveIPFSPath(p string, fullAccess bool, allowedDirs []string, resolve func(string) (string, error)) (string, error) {
	cleaned, err := cleanAbsFilePath(p)
	if err != nil {
		return "", err
	}
	if fullAccess {
		return cleaned, nil
	}
	return confineToAllowedDirs(cleaned, allowedDirs, resolve)
}

// cleanAbsFilePath returns the cleaned form of p when it is an absolute path
// that names a file: not a directory spelling ("/", a trailing separator, "."
// or ".."), and not a UNC path, which would make the host contact a remote
// server just to inspect it.
func cleanAbsFilePath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("%w: path is required", errIPFSPathInvalid)
	}
	cleaned := filepath.Clean(p)
	if !filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("%w: path must be absolute", errIPFSPathInvalid)
	}
	if strings.HasPrefix(filepath.VolumeName(cleaned), `\\`) {
		return "", fmt.Errorf("%w: UNC paths are not allowed", errIPFSPathInvalid)
	}
	if base := filepath.Base(p); base == "." || base == ".." ||
		strings.HasSuffix(p, "/") || strings.HasSuffix(p, string(filepath.Separator)) ||
		filepath.Base(cleaned) == string(filepath.Separator) {
		return "", fmt.Errorf("%w: path must name a file", errIPFSPathInvalid)
	}
	return cleaned, nil
}

// resolveParentDir resolves symlinks in the directory part of p and keeps the
// final component as given, for a file that does not exist yet.
func resolveParentDir(p string) (string, error) {
	dir, err := filepath.EvalSymlinks(filepath.Dir(p))
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filepath.Base(p)), nil
}

// confineToAllowedDirs returns the real path of candidate when it lies inside
// one of allowedDirs, and errIPFSPathNotAllowed otherwise. Symlinks are
// resolved on both sides before comparing (resolve for the candidate,
// EvalSymlinks for each allowed dir), so a link planted inside an allowed
// directory cannot redirect the access. Both the configured and the resolved
// spelling of each allowed dir count, so an operator may configure either
// (macOS: /var vs /private/var). The candidate is a local absolute path by the
// time it gets here: cleanAbsFilePath has already refused UNC paths, so the
// resolve step never reaches out to a remote host.
func confineToAllowedDirs(candidate string, allowedDirs []string, resolve func(string) (string, error)) (string, error) {
	if len(allowedDirs) == 0 {
		return "", fmt.Errorf("%w: no IPFS_ALLOWED_DIRS configured for users without full access", errIPFSPathNotAllowed)
	}
	roots := make([]string, 0, 2*len(allowedDirs))
	for _, dir := range allowedDirs {
		dir = filepath.Clean(dir)
		roots = append(roots, dir)
		if real, err := filepath.EvalSymlinks(dir); err == nil && real != dir {
			roots = append(roots, real)
		}
	}
	resolved, err := resolve(candidate)
	if err != nil {
		return "", fmt.Errorf("%w: path does not exist", errIPFSPathNotAllowed)
	}
	for _, root := range roots {
		if lib.PathUnderDir(resolved, root) {
			return resolved, nil
		}
	}
	return "", fmt.Errorf("%w: path is outside the configured IPFS directories", errIPFSPathNotAllowed)
}

// createDownloadFile opens the destination for writing with O_EXCL: never
// clobber an existing file. The destination is chosen by the API caller, so a
// plain Create let anyone holding ipfs_get overwrite proxy.conf, the cookie
// file or a shell profile with content they had pinned to IPFS.
func createDownloadFile(destinationPath string) (*os.File, error) {
	f, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("destination file already exists: %s", destinationPath)
		}
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	return f, nil
}

// ipfsPathErrorStatus maps a path validation error to an HTTP status.
func ipfsPathErrorStatus(err error) int {
	if errors.Is(err, errIPFSPathNotAllowed) {
		return http.StatusForbidden
	}
	return http.StatusBadRequest
}
