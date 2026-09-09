package files

import (
	"testing"

	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

func TestFileGuard(t *testing.T) {
	guard := NewGuard()
	p := core.GuardrailsConfig{
		Files: core.FilesPolicy{
			Protected: []string{
				".env*",
				"secrets/**",
				"**/*.pem",
				"**/*.key",
				"credentials.json",
			},
			Action: "deny",
		},
	}

	tests := []struct {
		name        string
		tool        string
		path        string
		expectBlock bool
	}{
		{".env write", "write_to_file", ".env", true},
		{".env.production write", "write_to_file", ".env.production", true},
		{"nested secrets write", "replace_file_content", "config/secrets/prod.json", true},
		{"private key replace", "replace_file_content", "certs/server.key", true},
		{"credentials edit", "edit_file", "credentials.json", true},
		{"normal code file", "write_to_file", "core/config.go", false},
		{"normal markdown", "replace_file_content", "README.md", false},
		{"read tool not blocked", "view_file", ".env", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := guard.Check(tt.tool, map[string]any{"TargetFile": tt.path}, p)
			if tt.expectBlock {
				if res == nil || res.Decision != api.DecisionDeny {
					t.Errorf("Expected path '%s' with tool '%s' to be blocked, got: %v", tt.path, tt.tool, res)
				}
			} else {
				if res != nil && res.IsBlocked() {
					t.Errorf("Expected path '%s' with tool '%s' to be allowed, got: %v", tt.path, tt.tool, res)
				}
			}
		})
	}
}
