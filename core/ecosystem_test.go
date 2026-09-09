package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectEcosystem(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Python detection
	pyproject := filepath.Join(tempDir, "pyproject.toml")
	_ = os.WriteFile(pyproject, []byte("[project]\nname='my-python-app'"), 0644)

	eco := DetectEcosystem(tempDir)
	if eco.TestCommand != "pytest" {
		t.Errorf("Expected test command 'pytest', got '%s'", eco.TestCommand)
	}
	foundPy := false
	for _, l := range eco.Languages {
		if l == "Python" {
			foundPy = true
		}
	}
	if !foundPy {
		t.Errorf("Expected Python in languages, got %v", eco.Languages)
	}

	// 2. TypeScript / Node detection
	tempDirTS := t.TempDir()
	pkgJSON := filepath.Join(tempDirTS, "package.json")
	_ = os.WriteFile(pkgJSON, []byte(`{"name":"web-ui","scripts":{"test":"vitest","typecheck":"tsc --noEmit","lint":"eslint ."}}`), 0644)
	tsconfig := filepath.Join(tempDirTS, "tsconfig.json")
	_ = os.WriteFile(tsconfig, []byte("{}"), 0644)

	ecoTS := DetectEcosystem(tempDirTS)
	if ecoTS.Name != "web-ui" {
		t.Errorf("Expected project name 'web-ui', got '%s'", ecoTS.Name)
	}
	if ecoTS.TestCommand != "npm test" {
		t.Errorf("Expected test command 'npm test', got '%s'", ecoTS.TestCommand)
	}
	if ecoTS.TypecheckCommand != "npm run typecheck" {
		t.Errorf("Expected typecheck command 'npm run typecheck', got '%s'", ecoTS.TypecheckCommand)
	}

	// 3. Polyglot monorepo detection (Go backend + Node frontend)
	tempDirMono := t.TempDir()
	backendDir := filepath.Join(tempDirMono, "backend")
	_ = os.MkdirAll(backendDir, 0755)
	_ = os.WriteFile(filepath.Join(backendDir, "go.mod"), []byte("module mybackend\n\ngo 1.22"), 0644)

	frontendDir := filepath.Join(tempDirMono, "frontend")
	_ = os.MkdirAll(frontendDir, 0755)
	_ = os.WriteFile(filepath.Join(frontendDir, "package.json"), []byte(`{"name":"client","scripts":{"test":"vitest"}}`), 0644)

	ecoMono := DetectEcosystem(tempDirMono)
	if len(ecoMono.Languages) < 2 {
		t.Errorf("Expected at least 2 languages in polyglot repo, got %v", ecoMono.Languages)
	}
}
