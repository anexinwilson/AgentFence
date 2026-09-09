package commands

import (
	"fmt"
	"os"

	"github.com/agentfence/agentfence/cli/ui"
	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/verification"
	"github.com/agentfence/agentfence/verification/attestation"
)

// RunVerify executes the comprehensive verification pipeline.
func RunVerify(args []string) {
	fmt.Print(ui.Banner())
	ui.Section("Running AgentFence Verification")

	cfg, err := core.Load("")
	if err != nil {
		ui.Fail("Failed to load policy: %v", err)
		os.Exit(1)
	}

	cwd, _ := os.Getwd()
	engine := verification.NewEngine()
	report := engine.Run(cfg.Guardrails, cwd)

	ui.PrintReport(report)

	if !report.AllPassed {
		os.Exit(1)
	}

	if att, err := attestation.Generate(cwd, "Verification passed: all checks green"); err == nil {
		ui.Pass("Attestation locked: .agentfence/attestation.json (tree: %s)", att.GitTreeSHA)
	}
}
