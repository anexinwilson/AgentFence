package enforcement

import (
	"testing"

	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

func TestRouter(t *testing.T) {
	router := NewRouter()
	cfg := core.DefaultConfig("test-project")

	// 1. Should block git push
	res := router.Evaluate("run_command", map[string]any{"CommandLine": "git push origin main"}, cfg.Guardrails)
	if res.Decision != api.DecisionDeny {
		t.Errorf("Expected git push to be denied, got: %s", res.Decision)
	}

	// 2. Should block rm -rf *
	res = router.Evaluate("run_command", map[string]any{"CommandLine": "rm -rf *"}, cfg.Guardrails)
	if res.Decision != api.DecisionDeny {
		t.Errorf("Expected rm -rf * to be denied, got: %s", res.Decision)
	}

	// 3. Should block writing to .env
	res = router.Evaluate("write_to_file", map[string]any{"TargetFile": ".env"}, cfg.Guardrails)
	if res.Decision != api.DecisionDeny {
		t.Errorf("Expected .env write to be denied, got: %s", res.Decision)
	}

	// 4. Should allow safe git status
	res = router.Evaluate("run_command", map[string]any{"CommandLine": "git status"}, cfg.Guardrails)
	if res.Decision != api.DecisionAllow {
		t.Errorf("Expected git status to be allowed, got: %s", res.Decision)
	}

	// 5. Should allow writing normal go files
	res = router.Evaluate("write_to_file", map[string]any{"TargetFile": "core/config.go"}, cfg.Guardrails)
	if res.Decision != api.DecisionAllow {
		t.Errorf("Expected config.go write to be allowed, got: %s", res.Decision)
	}
}
