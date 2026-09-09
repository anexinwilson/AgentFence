package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentfence/agentfence/cli/ui"
	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/evidence"
)

// RunEvidence compiles and presents the objective evidence report for the current task.
func RunEvidence(args []string) {
	asMarkdown := false
	asJSON := false

	for _, arg := range args {
		if arg == "--markdown" || arg == "-m" {
			asMarkdown = true
		} else if arg == "--json" || arg == "-j" {
			asJSON = true
		}
	}

	cfg, err := core.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load policy: %v\n", err)
		os.Exit(1)
	}

	cwd, _ := os.Getwd()
	collector := evidence.NewCollector()
	report := collector.Collect(cfg.Guardrails, cfg.Project.Name, cwd)

	mdReport := evidence.FormatEvidenceMarkdown(report)

	// Persist evidence report to .agents/evidence.md for a permanent audit trail
	agentsDir := filepath.Join(cwd, ".agents")
	if fi, err := os.Stat(agentsDir); err == nil && fi.IsDir() {
		_ = os.WriteFile(filepath.Join(agentsDir, "evidence.md"), []byte(mdReport), 0644)
	}

	if asJSON {
		bytes, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(bytes))
	} else if asMarkdown {
		fmt.Println(mdReport)
	} else {
		fmt.Print(ui.Banner())
		evidence.PrintEvidenceConsole(report)
	}

	if report.Verdict != "VERIFIED" {
		os.Exit(1)
	}
}
