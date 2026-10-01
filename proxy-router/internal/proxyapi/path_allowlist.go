package proxyapi

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

// DefaultFilePathAllowRoots returns roots under which HTTP client-supplied
// filesystem paths may be read (IPFS add, etc.).
func DefaultFilePathAllowRoots() []string {
	roots := []string{os.TempDir()}
	if cwd, err := os.Getwd(); err == nil {
		roots = append(roots, cwd)
	}
	return roots
}

// ValidateAllowlistedFilePath rejects empty paths and paths that do not stay
// under one of the allowlisted roots after Clean+Abs.
func ValidateAllowlistedFilePath(filePath string, roots []string) error {
	if filePath == "" {
		return fmt.Errorf("filePath required")
	}
	abs, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("invalid filePath: %w", err)
	}
	for _, root := range roots {
		if root == "" {
			continue
		}
		if lib.PathUnderDir(abs, root) {
			return nil
		}
	}
	return fmt.Errorf("filePath outside allowlisted roots")
}
