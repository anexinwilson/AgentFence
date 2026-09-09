package paths

import (
	"path/filepath"
	"strings"
)

// Normalize converts Windows backslashes to forward slashes and trims whitespace.
func Normalize(p string) string {
	return filepath.ToSlash(strings.TrimSpace(p))
}

// SubpathMatches checks if target matches a glob pattern across OS path separators.
func SubpathMatches(pattern, path string) bool {
	normPattern := strings.ToLower(Normalize(pattern))
	normPath := strings.ToLower(Normalize(path))

	if matched, _ := filepath.Match(normPattern, normPath); matched {
		return true
	}

	base := filepath.Base(normPath)
	if matched, _ := filepath.Match(normPattern, base); matched {
		return true
	}

	if strings.HasPrefix(normPattern, "**/") {
		suffix := strings.TrimPrefix(normPattern, "**/")
		if matched, _ := filepath.Match(suffix, base); matched {
			return true
		}
	}

	return false
}
