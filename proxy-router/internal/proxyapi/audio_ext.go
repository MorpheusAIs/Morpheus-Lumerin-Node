package proxyapi

import (
	"path/filepath"
	"strings"
)

// audioExtensionContentTypes is the allowlist of audio file extensions the
// proxy understands, mapped to their content types. It is the single source
// for both content-type detection and for deciding which client-provided
// extension may be preserved on a server-generated temp file name.
var audioExtensionContentTypes = map[string]string{
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".wave": "audio/wav",
	".ogg":  "audio/ogg",
	".flac": "audio/flac",
	".aac":  "audio/aac",
	".m4a":  "audio/mp4",
	".webm": "audio/webm",
	".opus": "audio/opus",
	".wma":  "audio/x-ms-wma",
	".amr":  "audio/amr",
	".3gp":  "audio/3gpp",
	".aiff": "audio/aiff",
}

// safeAudioExtension returns the lowercased extension of fileName if it is in
// the allowlist, otherwise "". Only the extension is ever taken from a client
// filename; the rest of the name is discarded.
func safeAudioExtension(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if _, ok := audioExtensionContentTypes[ext]; ok {
		return ext
	}
	return ""
}
