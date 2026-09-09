package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

// Guard prevents execution of dangerous shell commands.
type Guard struct{}

func NewGuard() *Guard {
	return &Guard{}
}

func (g *Guard) Name() string {
	return "commands"
}

func (g *Guard) Check(toolName string, toolArgs map[string]any, p core.GuardrailsConfig) *api.GuardResult {
	if toolName != "run_command" {
		return nil
	}

	commandLine, _ := extractCommand(toolArgs)
	if commandLine == "" {
		return nil
	}

	normCmd := strings.TrimSpace(strings.ToLower(commandLine))
	decision := api.DecisionDeny
	if strings.ToLower(p.Commands.Action) == "ask" {
		decision = api.DecisionAsk
	}

	for _, pattern := range p.Commands.Block {
		pattern = strings.TrimSpace(strings.ToLower(pattern))
		if pattern == "" {
			continue
		}

		if matched, _ := filepath.Match(pattern, normCmd); matched {
			return g.makeResult(decision, pattern, commandLine)
		}

		if strings.Contains(pattern, "|") {
			parts := strings.Split(pattern, "|")
			allFound := true
			for _, part := range parts {
				cleanPart := strings.TrimSpace(strings.ReplaceAll(part, "*", ""))
				if cleanPart != "" && !strings.Contains(normCmd, cleanPart) {
					allFound = false
					break
				}
			}
			if allFound {
				return g.makeResult(decision, pattern, commandLine)
			}
		}

		clean := strings.TrimSpace(strings.ReplaceAll(pattern, "*", ""))
		if clean != "" && strings.Contains(normCmd, clean) {
			return g.makeResult(decision, pattern, commandLine)
		}
	}

	return nil
}

func (g *Guard) makeResult(decision api.Decision, pattern, commandLine string) *api.GuardResult {
	return &api.GuardResult{
		Decision:  decision,
		Reason:    fmt.Sprintf("AgentFence blocked execution of prohibited command pattern '%s'. Command was: '%s'", pattern, commandLine),
		GuardName: g.Name(),
		RuleName:  fmt.Sprintf("command.%s", pattern),
	}
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
