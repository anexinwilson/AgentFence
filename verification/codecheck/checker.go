package codecheck

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

// Violation represents an AST integrity violation (hollow mock, empty stub, or fake container).
type Violation struct {
	FilePath string
	LineNum  int
	Snippet  string
	Reason   string
}

func (v Violation) String() string {
	if v.LineNum > 0 {
		return fmt.Sprintf("%s:%d: %s (%s)", v.FilePath, v.LineNum, v.Snippet, v.Reason)
	}
	return fmt.Sprintf("%s: %s (%s)", v.FilePath, v.Snippet, v.Reason)
}

// Validator inspects code AST and container configurations for hollow implementations.
type Validator struct{}

// NewValidator creates a new code and infrastructure integrity validator.
func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Name() string {
	return "integrity"
}

// Validate executes AST hollow-function scanning and Dockerfile packaging checks.
func (v *Validator) Validate(p core.GuardrailsConfig, repoRoot string) api.ValidationResult {
	if repoRoot == "" {
		cwd, _ := os.Getwd()
		repoRoot = cwd
	}

	var violations []Violation

	// 1. Scan Go AST
	goViolations := scanGoAST(repoRoot)
	violations = append(violations, goViolations...)

	// 2. Scan TypeScript/JavaScript/Python hollow stubs
	scriptViolations := scanScriptHollows(repoRoot)
	violations = append(violations, scriptViolations...)

	// 3. Scan Dockerfile OCI packaging
	dockerViolations := scanDockerIntegrity(repoRoot)
	violations = append(violations, dockerViolations...)

	if len(violations) == 0 {
		return api.ValidationResult{
			Name:    v.Name(),
			Passed:  true,
			Details: "Code AST and container packaging verified: zero hollow stubs or generic base containers detected.",
		}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Integrity check detected %d hollow or fake implementations:", len(violations)))
	for _, viol := range violations {
		lines = append(lines, fmt.Sprintf("  ✗ %s", viol.String()))
	}

	return api.ValidationResult{
		Name:          v.Name(),
		Passed:        false,
		Details:       strings.Join(lines, "\n"),
		ActionableFix: "Implement real business logic and ensure Dockerfiles package actual application source code.",
	}
}

func scanGoAST(repoRoot string) []Violation {
	var violations []Violation

	_ = filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && (info.Name() == "vendor" || info.Name() == ".git" || info.Name() == "testdata") {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel, _ := filepath.Rel(repoRoot, path)
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil
		}

		ast.Inspect(node, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				return true
			}

			// Empty body: func foo() {}
			if len(fn.Body.List) == 0 {
				pos := fset.Position(fn.Pos())
				violations = append(violations, Violation{
					FilePath: rel,
					LineNum:  pos.Line,
					Snippet:  fn.Name.Name + "() {}",
					Reason:   "Empty function body with zero implementation",
				})
				return true
			}

			// Single statement checks
			if len(fn.Body.List) == 1 {
				stmt := fn.Body.List[0]
				pos := fset.Position(stmt.Pos())

				// Check panic for stub invocation
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					if call, ok := exprStmt.X.(*ast.CallExpr); ok {
						if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
							violations = append(violations, Violation{
								FilePath: rel,
								LineNum:  pos.Line,
								Snippet:  "panic(...)",
								Reason:   "Unimplemented stub function (panic)",
							})
						}
					}
				}

				// Check hollow I/O functions that return literal nil/empty with no calls
				fnName := fn.Name.Name
				isIOFunction := strings.HasPrefix(fnName, "Fetch") ||
					strings.HasPrefix(fnName, "Query") ||
					strings.HasPrefix(fnName, "Send") ||
					strings.HasPrefix(fnName, "Load")

				if isIOFunction {
					if retStmt, ok := stmt.(*ast.ReturnStmt); ok && len(retStmt.Results) > 0 {
						hasCall := false
						ast.Inspect(retStmt, func(rn ast.Node) bool {
							if _, isCall := rn.(*ast.CallExpr); isCall {
								hasCall = true
							}
							return true
						})
						if !hasCall {
							violations = append(violations, Violation{
								FilePath: rel,
								LineNum:  pos.Line,
								Snippet:  fnName + " returns hardcoded literal without I/O",
								Reason:   "Hollow I/O function returning static literal without external calls",
							})
						}
					}
				}
			}

			return true
		})

		return nil
	})

	return violations
}

func scanScriptHollows(repoRoot string) []Violation {
	var violations []Violation

	targetExts := map[string]bool{
		".ts":  true,
		".tsx": true,
		".js":  true,
		".jsx": true,
		".py":  true,
	}

	_ = filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && (info.Name() == "node_modules" || info.Name() == ".git" || info.Name() == ".venv") {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if !targetExts[ext] {
			return nil
		}

		rel, _ := filepath.Rel(repoRoot, path)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(data), "\n")
		for i := 0; i < len(lines); i++ {
			trimmed := strings.TrimSpace(lines[i])

			// Python: def fetch_something(): pass / ...
			if ext == ".py" && strings.HasPrefix(trimmed, "def ") {
				for j := i + 1; j < len(lines) && j <= i+3; j++ {
					subTrimmed := strings.TrimSpace(lines[j])
					stubKeyword := "raise " + "NotImplemented" + "Error"
					if subTrimmed == "pass" || subTrimmed == "..." || subTrimmed == stubKeyword {
						violations = append(violations, Violation{
							FilePath: rel,
							LineNum:  j + 1,
							Snippet:  trimmed + " -> " + subTrimmed,
							Reason:   "Hollow Python function body (stub)",
						})
						break
					}
				}
			}

			// TypeScript/JS: async function fetch...() { return []; } with zero I/O
			if (ext == ".ts" || ext == ".tsx" || ext == ".js") && strings.HasPrefix(trimmed, "async function ") {
				if strings.Contains(trimmed, "fetch") || strings.Contains(trimmed, "query") || strings.Contains(trimmed, "load") {
					for j := i + 1; j < len(lines) && j <= i+4; j++ {
						subTrimmed := strings.TrimSpace(lines[j])
						if subTrimmed == "return [];" || subTrimmed == "return {};" || subTrimmed == "return null;" {
							violations = append(violations, Violation{
								FilePath: rel,
								LineNum:  j + 1,
								Snippet:  trimmed + " -> " + subTrimmed,
								Reason:   "Hollow async I/O function returning hardcoded literal with zero calls",
							})
							break
						}
					}
				}
			}
		}

		return nil
	})

	return violations
}

func scanDockerIntegrity(repoRoot string) []Violation {
	var violations []Violation

	dockerfiles := []string{"Dockerfile", "Dockerfile.dev", "Dockerfile.prod"}
	for _, dfName := range dockerfiles {
		dfPath := filepath.Join(repoRoot, dfName)
		data, err := os.ReadFile(dfPath)
		if err != nil {
			continue
		}

		content := string(data)
		hasCopy := false
		hasAdd := false

		lines := strings.Split(content, "\n")
		for _, line := range lines {
			upper := strings.ToUpper(strings.TrimSpace(line))
			if strings.HasPrefix(upper, "COPY ") {
				hasCopy = true
			}
			if strings.HasPrefix(upper, "ADD ") {
				hasAdd = true
			}
		}

		// Flag if Dockerfile pulls a base image but copies zero application files
		if !hasCopy && !hasAdd && len(lines) > 2 {
			violations = append(violations, Violation{
				FilePath: dfName,
				LineNum:  1,
				Snippet:  "Missing COPY/ADD instruction",
				Reason:   "Dockerfile does not package any project source files (generic base container stub)",
			})
		}
	}

	return violations
}
