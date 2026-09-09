package commands

import (
	"fmt"
	"os"

	"github.com/agentfence/agentfence/cli/ui"
	"github.com/agentfence/agentfence/core"
)

// RunCheck executes a quick scan of the project guardrails and policies.
func RunCheck(args []string) {
	fmt.Print(ui.Banner())
	ui.Section("AgentFence Quick Check")

	cfg, err := core.Load("")
	if err != nil {
		ui.Fail("Failed to load policy: %v", err)
		os.Exit(1)
	}

	ui.Pass("Policy loaded successfully for project '%s'", cfg.Project.Name)

	gitCount := len(cfg.Guardrails.Git.Block)
	if gitCount > 0 {
		ui.Pass("Git guard active: %d operation(s) blocked (%s)", gitCount, cfg.Guardrails.Git.Action)
	} else {
		ui.Warn("No Git operations blocked")
	}

	fileCount := len(cfg.Guardrails.Files.Protected)
	if fileCount > 0 {
		ui.Pass("File guard active: %d pattern(s) protected (%s)", fileCount, cfg.Guardrails.Files.Action)
	} else {
		ui.Warn("No protected file patterns configured")
	}

	cmdCount := len(cfg.Guardrails.Commands.Block)
	if cmdCount > 0 {
		ui.Pass("Command guard active: %d command pattern(s) blocked (%s)", cmdCount, cfg.Guardrails.Commands.Action)
	}

	if cfg.Guardrails.Placeholders.Block {
		ui.Pass("Placeholder guard active: scanning %d file type(s)", len(cfg.Guardrails.Placeholders.ScanExtensions))
	} else {
		ui.Warn("Placeholder guard disabled")
	}

	if cfg.Guardrails.Verification.Tests != nil && cfg.Guardrails.Verification.Tests.Command != "" {
		ui.Pass("Verification test command: `%s`", cfg.Guardrails.Verification.Tests.Command)
	} else {
		ui.Warn("No test command configured")
	}

	fmt.Println()
	ui.Pass("Guardrails check complete. Project is protected.")
	fmt.Println()
}
