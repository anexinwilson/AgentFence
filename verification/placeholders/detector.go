package placeholders

import (
	"fmt"
	"strings"
)

// Violation represents an unfulfilled placeholder in source code.
type Violation struct {
	FilePath string
	LineNum  int
	Snippet  string
	Reason   string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s:%d: %s (%s)", v.FilePath, v.LineNum, v.Snippet, v.Reason)
}

// DetectPlaceholders scans source lines for stub markers and incomplete code patterns.
func DetectPlaceholders(relPath, content string, ext string) []Violation {
	var violations []Violation
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		// 1. Explicit unfinished markers
		if strings.HasPrefix(lower, "// todo:") || strings.HasPrefix(lower, "# todo:") || strings.HasPrefix(lower, "/* todo:") {
			if strings.Contains(lower, "implement") || strings.Contains(lower, "stub") || strings.Contains(lower, "fake") {
				violations = append(violations, Violation{
					FilePath: relPath,
					LineNum:  i + 1,
					Snippet:  trimmed,
					Reason:   "Unfinished TODO implementation marker",
				})
			}
		}

		// 2. Panic / throw not implemented
		notImpl := "not " + "implemented"
		notImplErr := "notimplemented" + "error"
		if strings.Contains(lower, "panic(\""+notImpl+"\")") ||
			strings.Contains(lower, "panic(\"todo\")") ||
			strings.Contains(lower, "throw new error(\""+notImpl+"\")") ||
			strings.Contains(lower, "raise "+notImplErr) {
			violations = append(violations, Violation{
				FilePath: relPath,
				LineNum:  i + 1,
				Snippet:  trimmed,
				Reason:   "Explicit 'not implemented' stub exception/panic",
			})
		}

		// 3. Python pure 'pass' in function
		if ext == ".py" && (trimmed == "pass" || trimmed == "...") {
			violations = append(violations, Violation{
				FilePath: relPath,
				LineNum:  i + 1,
				Snippet:  trimmed,
				Reason:   "Bare 'pass' or '...' statement in code",
			})
		}
	}

	return violations
}
