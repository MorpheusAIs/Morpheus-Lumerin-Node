package lib

import (
	"path/filepath"
	"regexp"
	"strings"
)

var fileNameRegexp = regexp.MustCompile(`[^\w\-\.]`)

func SanitizeFilename(fileName string) string {
	fileName = strings.ToLower(fileName)
	return fileNameRegexp.ReplaceAllLiteralString(fileName, "_")
}

func PathUnderDir(candidate, dir string) bool {
	if candidate == "" || dir == "" {
		return false
	}
	absCandidate, err := filepath.Abs(filepath.Clean(candidate))
	if err != nil {
		return false
	}
	absDir, err := filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absDir, absCandidate)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	sep := string(filepath.Separator)
	return rel != ".." && !strings.HasPrefix(rel, ".."+sep)
}
