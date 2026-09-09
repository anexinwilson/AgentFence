package fallbacks

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

var ignoredDirectories = map[string]bool{
	".git":          true,
	"node_modules":  true,
	"vendor":        true,
	".venv":         true,
	"venv":          true,
	"dist":          true,
	"build":         true,
	"test":          true,
	"tests":         true,
	"__pycache__":   true,
	".pytest_cache": true,
}

// Scanner traverses the repository checking for unhandled swallowed errors and fallbacks.
type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "fallbacks"
}

func (s *Scanner) Validate(p core.GuardrailsConfig, repoRoot string) api.ValidationResult {
	// Enabled by default unless explicitly disabled (Block: false and Action: allow)
	if !p.Fallbacks.Block && strings.ToLower(p.Fallbacks.Action) == "allow" {
		return api.ValidationResult{
			Name:    s.Name(),
			Passed:  true,
			Details: "Fallback checking is disabled.",
		}
	}

	allowedExts := map[string]bool{
		".py":   true,
		".ts":   true,
		".tsx":  true,
		".js":   true,
		".jsx":  true,
		".go":   true,
	}
	if len(p.Fallbacks.ScanExtensions) > 0 {
		allowedExts = make(map[string]bool)
		for _, ext := range p.Fallbacks.ScanExtensions {
			allowedExts[strings.ToLower(ext)] = true
		}
	}

	var allViolations []Violation

	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			name := strings.ToLower(d.Name())
			if ignoredDirectories[name] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !allowedExts[ext] {
			return nil
		}

		// Skip test files
		base := strings.ToLower(d.Name())
		if strings.HasSuffix(base, "_test.go") || strings.HasPrefix(base, "test_") || strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(repoRoot, path)
		violations := DetectFallbacks(filepath.ToSlash(rel), string(data), ext)
		allViolations = append(allViolations, violations...)

		return nil
	})

	if err != nil {
		return api.ValidationResult{
			Name:    s.Name(),
			Passed:  false,
			Details: fmt.Sprintf("Failed to scan files for fallbacks: %v", err),
		}
	}

	if len(allViolations) == 0 {
		return api.ValidationResult{
			Name:    s.Name(),
			Passed:  true,
			Details: "No swallowed exceptions or silent fallbacks detected.",
		}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Detected %d swallowed exception(s) or silent fallback(s):", len(allViolations)))
	for i, v := range allViolations {
		if i >= 10 {
			lines = append(lines, fmt.Sprintf("  ... and %d more.", len(allViolations)-10))
			break
		}
		lines = append(lines, fmt.Sprintf("  • %s", v.String()))
	}

	passed := strings.ToLower(p.Fallbacks.Action) == "warn"
	return api.ValidationResult{
		Name:          s.Name(),
		Passed:        passed,
		Details:       strings.Join(lines, "\n"),
		ActionableFix: "Remove silent fallbacks and swallowed exceptions. Fail loudly or log errors explicitly instead of masking failures.",
	}
}
