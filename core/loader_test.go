package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPolicyLoader(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "agentfence.yaml")

	yamlContent := `
project:
  name: test-app

guardrails:
  git:
    block:
      - push
      - commit
    allow_readonly: true
    action: deny
  files:
    protected:
      - ".env*"
    action: deny
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Project.Name != "test-app" {
		t.Errorf("Expected project name 'test-app', got '%s'", cfg.Project.Name)
	}

	if len(cfg.Guardrails.Git.Block) != 2 {
		t.Errorf("Expected 2 blocked git ops, got %d", len(cfg.Guardrails.Git.Block))
	}

	if len(cfg.Guardrails.Files.Protected) != 1 {
		t.Errorf("Expected 1 protected file pattern, got %d", len(cfg.Guardrails.Files.Protected))
	}
}
