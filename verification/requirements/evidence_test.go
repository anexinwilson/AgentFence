package requirements

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentfence/agentfence/core"
)

func TestVerifyEvidence(t *testing.T) {
	tempDir := t.TempDir()

	evidenceFile := filepath.Join(tempDir, "feature.go")
	if err := os.WriteFile(evidenceFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to create test evidence file: %v", err)
	}

	// 1. Existing non-empty file should satisfy requirement
	itemValid := core.RequirementItem{
		ID:           "R-01",
		Title:        "Feature implementation",
		EvidenceFile: "feature.go",
	}
	status := VerifyEvidence(itemValid, tempDir)
	if !status.Satisfied {
		t.Errorf("Expected requirement R-01 to be satisfied, reason: %s", status.Reason)
	}

	// 2. Non-existent file should fail
	itemMissing := core.RequirementItem{
		ID:           "R-02",
		Title:        "Missing feature",
		EvidenceFile: "nonexistent.go",
	}
	status = VerifyEvidence(itemMissing, tempDir)
	if status.Satisfied {
		t.Errorf("Expected requirement R-02 to fail due to missing file")
	}

	// 3. Empty file (0 bytes) should fail
	emptyFile := filepath.Join(tempDir, "empty.go")
	_ = os.WriteFile(emptyFile, []byte(""), 0644)
	itemEmpty := core.RequirementItem{
		ID:           "R-03",
		Title:        "Empty file",
		EvidenceFile: "empty.go",
	}
	status = VerifyEvidence(itemEmpty, tempDir)
	if status.Satisfied {
		t.Errorf("Expected requirement R-03 to fail due to empty file")
	}
}
