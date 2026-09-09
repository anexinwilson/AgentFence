package git

import (
	"fmt"
	"strings"

	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

// Guard enforces the Git policy on shell commands.
type Guard struct{}

func NewGuard() *Guard {
	return &Guard{}
}

func (g *Guard) Name() string {
	return "git"
}

func (g *Guard) Check(toolName string, toolArgs map[string]any, p core.GuardrailsConfig) *api.GuardResult {
	if toolName != "run_command" {
		return nil
	}

	commandLine, _ := extractCommand(toolArgs)
	if commandLine == "" {
		return nil
	}

	invocations := ParseInvocations(commandLine)
	if len(invocations) == 0 {
		return nil
	}

	decision := api.DecisionDeny
	if strings.ToLower(p.Git.Action) == "ask" {
		decision = api.DecisionAsk
	}

	for _, inv := range invocations {
		if p.Git.AllowReadOnly && inv.IsReadOnly() {
			continue
		}

		for _, blockedPattern := range p.Git.Block {
			if matchesBlocked(inv, blockedPattern) {
				return &api.GuardResult{
					Decision:  decision,
					Reason:    fmt.Sprintf("AgentFence blocked Git operation 'git %s'. Guardrail prohibits '%s'. Command was: '%s'", inv.Subcommand, blockedPattern, commandLine),
					GuardName: g.Name(),
					RuleName:  fmt.Sprintf("git.%s", blockedPattern),
				}
			}
		}
	}

	return nil
}

func extractCommand(args map[string]any) (string, bool) {
	if val, ok := args["CommandLine"].(string); ok && val != "" {
		return val, true
	}
	if val, ok := args["command"].(string); ok && val != "" {
		return val, true
	}
	return "", false
}

func matchesBlocked(inv Invocation, blocked string) bool {
	parts := strings.Fields(strings.TrimSpace(blocked))
	if len(parts) == 0 {
		return false
	}

	if strings.ToLower(parts[0]) != inv.Subcommand {
		return false
	}

	if len(parts) == 1 {
		return true
	}

	requiredFlags := parts[1:]
	for _, req := range requiredFlags {
		found := false
		for _, arg := range inv.FullArgs {
			if strings.EqualFold(arg, req) {
				found = true
				break
			}
			if strings.HasPrefix(req, "-") && !strings.HasPrefix(req, "--") && len(req) == 2 {
				if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && strings.ContainsRune(arg, rune(req[1])) {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}

	return true
}
