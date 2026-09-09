package codecheck

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentfence/agentfence/core"
)

func TestIntegrityValidator(t *testing.T) {
	tempDir := t.TempDir()

	// Initially empty: should pass
	v := NewValidator()
	cfg := core.DefaultConfig("test-app")
	res := v.Validate(cfg.Guardrails, tempDir)
	if !res.Passed {
		t.Fatalf("expected initial empty directory to pass integrity check, got: %s", res.Details)
	}

	// Add a hollow Go function with empty body
	hollowGo := filepath.Join(tempDir, "hollow.go")
	_ = os.WriteFile(hollowGo, []byte("package main\n\nfunc Unfinished() {}\n"), 0644)

	resFail := v.Validate(cfg.Guardrails, tempDir)
	if resFail.Passed {
		t.Fatalf("expected integrity check to fail on empty function body")
	}

	// Remove hollow file
	_ = os.Remove(hollowGo)

	// Add a generic Dockerfile stub without COPY or ADD
	dfPath := filepath.Join(tempDir, "Dockerfile")
	_ = os.WriteFile(dfPath, []byte("FROM alpine:latest\nRUN echo hi\nCMD [\"sh\"]\n"), 0644)

	resDF := v.Validate(cfg.Guardrails, tempDir)
	if resDF.Passed {
		t.Fatalf("expected integrity check to fail on generic base Dockerfile stub")
	}
}
