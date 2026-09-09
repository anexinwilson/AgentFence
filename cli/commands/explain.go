package commands

import (
	"fmt"
	"strings"

	"github.com/agentfence/agentfence/cli/ui"
	"github.com/agentfence/agentfence/core"
)

// RunExplain provides clear context on why an action is blocked and how to configure it.
func RunExplain(args []string) {
	fmt.Print(ui.Banner())
	ui.Section("AgentFence Policy Explanation")

	ruleQuery := ""
	if len(args) > 0 {
		ruleQuery = strings.ToLower(args[0])
	}

	cfg, _ := core.Load("")

	if ruleQuery == "" || strings.Contains(ruleQuery, "git") {
		fmt.Printf("%sRule: git.block%s\n", ui.Bold, ui.Reset)
		fmt.Printf("Action: %s\n", cfg.Guardrails.Git.Action)
		fmt.Printf("Blocked Operations: %v\n", cfg.Guardrails.Git.Block)
		fmt.Println("Reason: Prohibits AI agents from executing disruptive Git operations (e.g. push, commit, reset --hard).")
		fmt.Println("To customize: Edit guardrails.git in agentfence.yaml.")
		fmt.Println()
	}

	if ruleQuery == "" || strings.Contains(ruleQuery, "file") {
		fmt.Printf("%sRule: files.protected%s\n", ui.Bold, ui.Reset)
		fmt.Printf("Action: %s\n", cfg.Guardrails.Files.Action)
		fmt.Printf("Protected Patterns: %v\n", cfg.Guardrails.Files.Protected)
		fmt.Println("Reason: Prevents accidental overwrite of secret credentials, keys, and environment variables.")
		fmt.Println("To customize: Edit guardrails.files.protected in agentfence.yaml.")
		fmt.Println()
	}

	if ruleQuery == "" || strings.Contains(ruleQuery, "placeholder") {
		fmt.Printf("%sRule: placeholders.block%s\n", ui.Bold, ui.Reset)
		fmt.Printf("Action: %s\n", cfg.Guardrails.Placeholders.Action)
		fmt.Println("Reason: Prevents AI agents from claiming completion when code contains empty functions, 'pass', or 'TODO' stubs.")
		fmt.Println("To customize: Edit guardrails.placeholders in agentfence.yaml.")
		fmt.Println()
	}
}
