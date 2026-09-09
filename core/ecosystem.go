package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ProjectEcosystem represents the detected tech stack of a repository.
type ProjectEcosystem struct {
	Name             string
	Languages        []string
	TestCommand      string
	TypecheckCommand string
	LintCommand      string
	BuildCommand     string
	HasDocker        bool
}

var ignoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".venv":        true,
	"venv":         true,
	"bin":          true,
	"target":       true,
	"dist":         true,
	"build":        true,
	".next":        true,
	".agents":      true,
}

// DetectEcosystem inspects the repository root and immediate subdirectories
// to auto-detect polyglot tech stacks and appropriate toolchain commands.
func DetectEcosystem(repoRoot string) ProjectEcosystem {
	if repoRoot == "" {
		repoRoot, _ = os.Getwd()
	}

	name := filepath.Base(repoRoot)
	langMap := make(map[string]bool)
	var testCmds, typecheckCmds, lintCmds, buildCmds []string
	hasDocker := false

	// Helper to add command without duplicates
	appendCmd := func(list *[]string, cmd string) {
		if cmd == "" {
			return
		}
		for _, c := range *list {
			if c == cmd {
				return
			}
		}
		*list = append(*list, cmd)
	}

	// 1. Scan directory targets: root + depth 1 subdirectories
	var scanDirs []string
	scanDirs = append(scanDirs, repoRoot)

	entries, err := os.ReadDir(repoRoot)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() && !ignoredDirs[e.Name()] && !strings.HasPrefix(e.Name(), ".") {
				scanDirs = append(scanDirs, filepath.Join(repoRoot, e.Name()))
			}
		}
	}

	for _, dir := range scanDirs {
		rel, _ := filepath.Rel(repoRoot, dir)
		isSubdir := rel != "." && rel != ""

		prefix := ""
		if isSubdir {
			prefix = filepath.ToSlash(rel)
		}

		// Node / TS
		pkgJSONPath := filepath.Join(dir, "package.json")
		if data, err := os.ReadFile(pkgJSONPath); err == nil {
			var pkg struct {
				Name    string            `json:"name"`
				Scripts map[string]string `json:"scripts"`
			}
			_ = json.Unmarshal(data, &pkg)
			if !isSubdir && pkg.Name != "" {
				name = pkg.Name
			}

			hasTS := false
			if _, err := os.Stat(filepath.Join(dir, "tsconfig.json")); err == nil {
				hasTS = true
				langMap["TypeScript"] = true
			} else {
				langMap["JavaScript"] = true
			}

			if pkg.Scripts != nil {
				if _, ok := pkg.Scripts["test"]; ok {
					if isSubdir {
						appendCmd(&testCmds, "cd "+prefix+" && npm test")
					} else {
						appendCmd(&testCmds, "npm test")
					}
				}
				if _, ok := pkg.Scripts["typecheck"]; ok {
					if isSubdir {
						appendCmd(&typecheckCmds, "cd "+prefix+" && npm run typecheck")
					} else {
						appendCmd(&typecheckCmds, "npm run typecheck")
					}
				} else if hasTS {
					if isSubdir {
						appendCmd(&typecheckCmds, "cd "+prefix+" && npx tsc --noEmit")
					} else {
						appendCmd(&typecheckCmds, "npx tsc --noEmit")
					}
				}
				if _, ok := pkg.Scripts["lint"]; ok {
					if isSubdir {
						appendCmd(&lintCmds, "cd "+prefix+" && npm run lint")
					} else {
						appendCmd(&lintCmds, "npm run lint")
					}
				}
				if _, ok := pkg.Scripts["build"]; ok {
					if isSubdir {
						appendCmd(&buildCmds, "cd "+prefix+" && npm run build")
					} else {
						appendCmd(&buildCmds, "npm run build")
					}
				}
			} else if hasTS {
				if isSubdir {
					appendCmd(&typecheckCmds, "cd "+prefix+" && npx tsc --noEmit")
				} else {
					appendCmd(&typecheckCmds, "npx tsc --noEmit")
				}
			}
		}

		// Python
		hasPy := false
		if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
			hasPy = true
		} else if _, err := os.Stat(filepath.Join(dir, "requirements.txt")); err == nil {
			hasPy = true
		} else if _, err := os.Stat(filepath.Join(dir, "setup.py")); err == nil {
			hasPy = true
		}

		if hasPy {
			langMap["Python"] = true
			if isSubdir {
				appendCmd(&testCmds, "pytest "+prefix)
			} else {
				appendCmd(&testCmds, "pytest")
			}
		}

		// Go
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			langMap["Go"] = true
			if isSubdir {
				appendCmd(&testCmds, "cd "+prefix+" && go test ./...")
				appendCmd(&buildCmds, "cd "+prefix+" && go build ./...")
			} else {
				appendCmd(&testCmds, "go test ./...")
				appendCmd(&buildCmds, "go build ./...")
			}
		}

		// Rust
		if _, err := os.Stat(filepath.Join(dir, "Cargo.toml")); err == nil {
			langMap["Rust"] = true
			if isSubdir {
				appendCmd(&testCmds, "cargo test --manifest-path "+filepath.Join(prefix, "Cargo.toml"))
				appendCmd(&typecheckCmds, "cargo check --manifest-path "+filepath.Join(prefix, "Cargo.toml"))
			} else {
				appendCmd(&testCmds, "cargo test")
				appendCmd(&typecheckCmds, "cargo check")
			}
		}

		// Java / Kotlin
		if _, err := os.Stat(filepath.Join(dir, "pom.xml")); err == nil {
			langMap["Java"] = true
			appendCmd(&testCmds, "mvn test")
		} else if _, err := os.Stat(filepath.Join(dir, "build.gradle")); err == nil {
			langMap["Gradle"] = true
			appendCmd(&testCmds, "gradle test")
		}

		// Makefile / C
		if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil && len(testCmds) == 0 {
			langMap["Makefile"] = true
		}

		// Docker
		if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err == nil {
			hasDocker = true
		}
	}

	var languages []string
	for l := range langMap {
		languages = append(languages, l)
	}
	if hasDocker {
		languages = append(languages, "Docker")
	}
	if len(languages) == 0 {
		languages = append(languages, "Generic")
	}

	return ProjectEcosystem{
		Name:             name,
		Languages:        languages,
		TestCommand:      strings.Join(testCmds, "; "),
		TypecheckCommand: strings.Join(typecheckCmds, "; "),
		LintCommand:      strings.Join(lintCmds, "; "),
		BuildCommand:     strings.Join(buildCmds, "; "),
		HasDocker:        hasDocker,
	}
}

// Summary returns a human-readable summary of detected languages.
func (e ProjectEcosystem) Summary() string {
	return strings.Join(e.Languages, ", ")
}
