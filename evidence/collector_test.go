package evidence

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentfence/agentfence/core"
)

func TestCollector(t *testing.T) {
	tempDir := t.TempDir()

	evidencePath := filepath.Join(tempDir, "auth.go")
	if err := os.WriteFile(evidencePath, []byte("package main\nfunc Auth() {}\n"), 0644); err != nil {
		t.Fatalf("Failed to write evidence file: %v", err)
	}

	p := core.GuardrailsConfig{
		Requirements: core.RequirementsPolicy{
			Checklist: []core.RequirementItem{
				{
					ID:           "R-01",
					Title:        "Authentication implementation",
					EvidenceFile: "auth.go",
					Required:     true,
				},
			},
		},
		Verification: core.VerificationPolicy{},
	}

	collector := NewCollector()
	report := collector.Collect(p, "test-project", tempDir)

	if report.Verdict != "VERIFIED" {
		t.Errorf("Expected verdict 'VERIFIED', got '%s'", report.Verdict)
	}

	if len(report.Requirements) != 1 || !report.Requirements[0].Satisfied {
		t.Errorf("Expected requirement R-01 to be satisfied, got: %v", report.Requirements)
	}

	md := FormatEvidenceMarkdown(report)
	if !strings.Contains(md, "VERDICT: VERIFIED") {
		t.Errorf("Expected markdown output to contain 'VERDICT: VERIFIED', got:\n%s", md)
	}
	if !strings.Contains(md, "Requirements") {
		t.Errorf("Expected markdown output to contain 'Requirements', got:\n%s", md)
	}
}

func TestCollectorFailure(t *testing.T) {
	tempDir := t.TempDir()

	p := core.GuardrailsConfig{
		Requirements: core.RequirementsPolicy{
			Checklist: []core.RequirementItem{
				{
					ID:           "R-01",
					Title:        "Missing feature",
					EvidenceFile: "missing.go",
					Required:     true,
				},
			},
		},
	}

	collector := NewCollector()
	report := collector.Collect(p, "test-project", tempDir)

	if report.Verdict != "FAILED" {
		t.Errorf("Expected verdict 'FAILED' due to missing evidence, got '%s'", report.Verdict)
	}
}
