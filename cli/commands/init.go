package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentfence/agentfence/adapters/protocol"
	"github.com/agentfence/agentfence/cli/ui"
	"github.com/agentfence/agentfence/core"
)

// RunInit initializes AgentFence policy and Antigravity hooks in the target repository.
func RunInit(args []string) {
	fmt.Print(ui.Banner())
	ui.Section("AgentFence Initialization")

	cwd, err := os.Getwd()
	if err != nil {
		ui.Fail("Failed to determine current working directory: %v", err)
		os.Exit(1)
	}

	projectName := filepath.Base(cwd)

	// 1. Detect Git
	gitDir := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		ui.Pass("Git repository detected")
	} else {
		ui.Warn("No .git directory found (Git policy guardrails will still apply if git is invoked)")
	}

	// 2. Create agentfence.yaml
	configPath := filepath.Join(cwd, "agentfence.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		content := core.GenerateDefaultYAML(projectName, cwd)
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			ui.Fail("Failed to write %s: %v", configPath, err)
			os.Exit(1)
		}
		ui.Pass("Created agentfence.yaml")
	} else {
		ui.Info("agentfence.yaml already exists (preserving)")
	}

	// 3. Create .agents directory and hooks.json
	agentsDir := filepath.Join(cwd, ".agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		ui.Fail("Failed to create .agents directory: %v", err)
		os.Exit(1)
	}

	hooksPath := filepath.Join(agentsDir, "hooks.json")
	hooksJSON, err := protocol.GenerateHooksJSON("agentfence")
	if err != nil {
		ui.Fail("Failed to generate hooks.json: %v", err)
		os.Exit(1)
	}

	if err := os.WriteFile(hooksPath, []byte(hooksJSON), 0644); err != nil {
		ui.Fail("Failed to write %s: %v", hooksPath, err)
		os.Exit(1)
	}
	ui.Pass("Created .agents/hooks.json (configured PreToolUse & Stop hooks)")

	// 4. Create .agents/rules/agentfence.md
	rulesDir := filepath.Join(agentsDir, "rules")
	_ = os.MkdirAll(rulesDir, 0755)
	rulePath := filepath.Join(rulesDir, "agentfence.md")

	ruleContent := `# AgentFence Guardrails & Architectural Laws

This project is protected by AgentFence.
The AI agent must strictly adhere to these laws:

1. **Do not perform prohibited Git operations**:
   Operations such as git push, git commit, git reset --hard, or git clean -f are strictly intercepted and denied.
2. **Do not modify protected files**:
   Secrets, credentials, and .env files cannot be directly edited or overwritten.
3. **No silent fallbacks / No dual systems**:
   Never catch an exception just to return mock data, fake responses, or a fallback pipeline. If a system fails, fail loudly so the real root cause is visible. We do not build or maintain dual systems.
4. **No overengineering**:
   Implement the direct, standard solution. Do not introduce artificial sliding windows, synthetic status milestones, or extra abstraction wrappers. Keep files focused and adhere to single responsibility.
5. **No placeholders or silent substitutions**:
   Do not leave functions with bare 'pass', '...', 'NotImplementedError', or unfulfilled TODOs. Never substitute precompiled artifacts, cached models, or empty container images when a build is requested.
6. **Deterministic verification is required**:
   The agent will not be permitted to finish (Stop hook) until tests pass, typechecks pass, and required evidence is verified.
7. **Objective evidence reporting**:
   At the conclusion of the task, the agent must run 'agentfence evidence --markdown' and display the clean, objective Evidence Report. Never self-certify completion with unverified verbal claims.
`
	if err := os.WriteFile(rulePath, []byte(ruleContent), 0644); err != nil {
		ui.Fail("Failed to write %s: %v", rulePath, err)
		os.Exit(1)
	}
	ui.Pass("Created .agents/rules/agentfence.md")

	fmt.Println()
	ui.Pass("AgentFence is initialized and ready.")
	fmt.Printf("\nNext steps:\n  agentfence check    # Perform quick guardrail check\n  agentfence verify   # Run full verification suite\n  agentfence evidence # Generate task proof\n\n")
}
