package files

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

var fileMutatingTools = map[string]bool{
	"write_to_file":        true,
	"replace_file_content": true,
	"edit_file":            true,
	"delete_file":          true,
}

// Guard prevents modification of protected files.
type Guard struct{}

func NewGuard() *Guard {
	return &Guard{}
}

func (g *Guard) Name() string {
	return "files"
}

func (g *Guard) Check(toolName string, toolArgs map[string]any, p core.GuardrailsConfig) *api.GuardResult {
	if !fileMutatingTools[toolName] {
		return nil
	}

	targetPath := extractPath(toolArgs)
	if targetPath == "" {
		return nil
	}

	normalized := filepath.ToSlash(targetPath)
	base := filepath.Base(normalized)

	decision := api.DecisionDeny
	if strings.ToLower(p.Files.Action) == "ask" {
		decision = api.DecisionAsk
	}

	for _, pattern := range p.Files.Protected {
		pattern = strings.TrimSpace(filepath.ToSlash(pattern))
		if pattern == "" {
			continue
		}

		if matched, _ := filepath.Match(strings.ToLower(pattern), strings.ToLower(base)); matched {
			return g.makeResult(decision, pattern, targetPath)
		}

		if matched, _ := filepath.Match(strings.ToLower(pattern), strings.ToLower(normalized)); matched {
			return g.makeResult(decision, pattern, targetPath)
		}

		if strings.HasPrefix(pattern, "**/") {
			suffix := strings.TrimPrefix(pattern, "**/")
			if matched, _ := filepath.Match(strings.ToLower(suffix), strings.ToLower(base)); matched {
				return g.makeResult(decision, pattern, targetPath)
			}
		}

		if strings.Contains(strings.ToLower(normalized), strings.ToLower(strings.TrimSuffix(pattern, "**"))) {
			return g.makeResult(decision, pattern, targetPath)
		}
	}

	return nil
}

func (g *Guard) makeResult(decision api.Decision, pattern, targetPath string) *api.GuardResult {
	return &api.GuardResult{
		Decision:  decision,
		Reason:    fmt.Sprintf("AgentFence blocked modification of protected file '%s'. Matched guardrail pattern: '%s'", targetPath, pattern),
		GuardName: g.Name(),
		RuleName:  fmt.Sprintf("file.%s", pattern),
	}
}

func extractPath(args map[string]any) string {
	keys := []string{"TargetFile", "target_file", "AbsolutePath", "file_path", "path"}
	for _, k := range keys {
		if val, ok := args[k].(string); ok && val != "" {
			return val
		}
	}
	return ""
}
