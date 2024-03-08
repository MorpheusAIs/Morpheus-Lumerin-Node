package lib

import (
	"regexp"
	"strings"
)

var fileNameRegexp = regexp.MustCompile(`[^\w\-\.]`)

// SanitizeFileName modifies filename to be safe to use as cross-platform file name
func SanitizeFilename(fileName string) string {
	fileName = strings.ToLower(fileName)
	return fileNameRegexp.ReplaceAllLiteralString(fileName, "_")
}
