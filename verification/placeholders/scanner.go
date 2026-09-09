package placeholders

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

// Scanner traverses the repository checking for unfulfilled placeholders.
type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "placeholders"
}

func (s *Scanner) Validate(p core.GuardrailsConfig, repoRoot string) api.ValidationResult {
	if !p.Placeholders.Block {
		return api.ValidationResult{
			Name:    s.Name(),
			Passed:  true,
			Details: "Placeholder checking is disabled.",
		}
	}

	allowedExts := make(map[string]bool)
	for _, ext := range p.Placeholders.ScanExtensions {
		allowedExts[strings.ToLower(ext)] = true
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
		if strings.HasSuffix(base, "_test.go") || strings.HasPrefix(base, "test_") || strings.Contains(base, ".test.") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(repoRoot, path)
		violations := DetectPlaceholders(filepath.ToSlash(rel), string(data), ext)
		allViolations = append(allViolations, violations...)

		return nil
	})

	if err != nil {
		return api.ValidationResult{
			Name:    s.Name(),
			Passed:  false,
			Details: fmt.Sprintf("Failed to scan files for placeholders: %v", err),
		}
	}

	if len(allViolations) == 0 {
		return api.ValidationResult{
			Name:    s.Name(),
			Passed:  true,
			Details: "No placeholder or stub implementations detected.",
		}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Detected %d placeholder or stub implementation(s):", len(allViolations)))
	for i, v := range allViolations {
		if i >= 10 {
			lines = append(lines, fmt.Sprintf("  ... and %d more.", len(allViolations)-10))
			break
		}
		lines = append(lines, fmt.Sprintf("  • %s", v.String()))
	}

	passed := strings.ToLower(p.Placeholders.Action) != "fail"
	return api.ValidationResult{
		Name:          s.Name(),
		Passed:        passed,
		Details:       strings.Join(lines, "\n"),
		ActionableFix: "Replace stub implementations, fake return values, and TODO comments with real code.",
	}
}
