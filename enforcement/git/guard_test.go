package git

import (
	"testing"

	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

func TestGitGuard(t *testing.T) {
	guard := NewGuard()
	p := core.GuardrailsConfig{
		Git: core.GitPolicy{
			Block:         []string{"push", "commit", "reset --hard", "clean -f", "rebase"},
			AllowReadOnly: true,
			Action:        "deny",
		},
	}

	tests := []struct {
		name        string
		command     string
		expectBlock bool
	}{
		{"push simple", "git push", true},
		{"push with remote and branch", "git push origin main", true},
		{"push with force", "git push --force origin main", true},
		{"commit with message", "git commit -m 'feat: test'", true},
		{"reset hard", "git reset --hard HEAD~1", true},
		{"clean force", "git clean -fd", true},
		{"rebase", "git rebase main", true},
		{"chained push", "git add . && git push", true},
		{"status (readonly)", "git status", false},
		{"diff (readonly)", "git diff HEAD", false},
		{"log (readonly)", "git log -n 5", false},
		{"show (readonly)", "git show HEAD", false},
		{"non-git command", "npm test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := guard.Check("run_command", map[string]any{"CommandLine": tt.command}, p)
			if tt.expectBlock {
				if res == nil || res.Decision != api.DecisionDeny {
					t.Errorf("Expected command '%s' to be blocked (deny), got: %v", tt.command, res)
				}
			} else {
				if res != nil && res.IsBlocked() {
					t.Errorf("Expected command '%s' to be allowed, got blocked: %v", tt.command, res)
				}
			}
		})
	}
}
