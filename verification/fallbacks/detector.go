package fallbacks

import (
	"fmt"
	"strings"
)

// Violation represents a silent fallback or swallowed error pattern.
type Violation struct {
	FilePath string
	LineNum  int
	Snippet  string
	Reason   string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s:%d: %s (%s)", v.FilePath, v.LineNum, v.Snippet, v.Reason)
}

// DetectFallbacks scans source code lines for swallowed exceptions, empty catch blocks,
// and silent fallback returns that mask failures and create dual-system maintenance debt.
func DetectFallbacks(relPath, content string, ext string) []Violation {
	var violations []Violation
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		lower := strings.ToLower(line)

		// 1. Python: Exception swallowing and fallback returns
		if ext == ".py" {
			if strings.HasPrefix(lower, "except") && strings.HasSuffix(lower, ":") {
				isBroadCatch := strings.Contains(lower, "exception") || strings.Contains(lower, "baseexception") || lower == "except:"
				
				// Look ahead to next non-empty line
				for j := i + 1; j < len(lines) && j <= i+3; j++ {
					nextLine := strings.TrimSpace(lines[j])
					if nextLine == "" || strings.HasPrefix(nextLine, "#") {
						continue
					}
					nextLower := strings.ToLower(nextLine)

					if nextLower == "pass" || nextLower == "..." {
						violations = append(violations, Violation{
							FilePath: relPath,
							LineNum:  j + 1,
							Snippet:  nextLine,
							Reason:   "Swallowed exception with bare 'pass'/'...' (fail loudly instead of silencing errors)",
						})
						break
					}

					if isBroadCatch && (strings.HasPrefix(nextLower, "return mock") || strings.HasPrefix(nextLower, "return fallback")) {
						violations = append(violations, Violation{
							FilePath: relPath,
							LineNum:  j + 1,
							Snippet:  nextLine,
							Reason:   "Silent fallback return in exception handler (creates unmaintainable dual systems)",
						})
						break
					}
					break
				}
			}
		}

		// 2. JavaScript / TypeScript: Empty catch blocks and mock/fallback returns
		if ext == ".ts" || ext == ".tsx" || ext == ".js" || ext == ".jsx" {
			// Single-line empty catch: catch (e) {}, catch {}
			if (strings.Contains(lower, "catch") && (strings.Contains(lower, "{}") || strings.Contains(lower, "{ }"))) {
				violations = append(violations, Violation{
					FilePath: relPath,
					LineNum:  i + 1,
					Snippet:  line,
					Reason:   "Empty catch block silently swallowing error (fail loudly or handle properly)",
				})
			}

			// Multi-line empty catch or fallback return
			if strings.HasPrefix(lower, "catch") && strings.HasSuffix(lower, "{") {
				for j := i + 1; j < len(lines) && j <= i+3; j++ {
					nextLine := strings.TrimSpace(lines[j])
					if nextLine == "" || strings.HasPrefix(nextLine, "//") {
						continue
					}
					nextLower := strings.ToLower(nextLine)

					if nextLower == "}" {
						violations = append(violations, Violation{
							FilePath: relPath,
							LineNum:  j + 1,
							Snippet:  "}",
							Reason:   "Empty catch block silently swallowing error",
						})
						break
					}

					if strings.Contains(nextLower, "return mock") || strings.Contains(nextLower, "return fallback") {
						violations = append(violations, Violation{
							FilePath: relPath,
							LineNum:  j + 1,
							Snippet:  nextLine,
							Reason:   "Silent fallback return in catch block (masks whether real system is working)",
						})
						break
					}
					break
				}
			}
		}

		// 3. Go: Discarded errors in explicit error checks
		if ext == ".go" {
			if strings.HasPrefix(lower, "if err != nil {") && strings.HasSuffix(lower, "}") {
				// Single line if err != nil {}
				violations = append(violations, Violation{
					FilePath: relPath,
					LineNum:  i + 1,
					Snippet:  line,
					Reason:   "Empty error block discarding error without handling or logging",
				})
			}
		}
	}

	return violations
}
