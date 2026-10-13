package config

import "os"

// syncContentFile writes content to path when the env value is non-empty and
// the file is missing or its bytes differ. An empty content leaves the file
// alone. wrote is true only when the file was created or replaced.
func syncContentFile(path, content string, perm os.FileMode) (wrote bool, err error) {
	if content == "" {
		return false, nil
	}
	existing, readErr := os.ReadFile(path)
	if readErr == nil && string(existing) == content {
		return false, nil
	}
	if readErr != nil && !os.IsNotExist(readErr) {
		return false, readErr
	}
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		return false, err
	}
	return true, nil
}
