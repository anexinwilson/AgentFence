package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentfence/agentfence/adapters/protocol"
	"github.com/agentfence/agentfence/cli/ui"
)

// RunStop temporarily pauses/disables AgentFence hooks and pre-commit checks.
func RunStop(args []string) {
	fmt.Print(ui.Banner())
	ui.Section("Stopping AgentFence Protection")

	cwd, err := os.Getwd()
	if err != nil {
		ui.Fail("Failed to determine current working directory: %v", err)
		os.Exit(1)
	}

	stoppedSomething := false

	// 1. Disable Antigravity hooks.json
	hooksPath := filepath.Join(cwd, ".agents", "hooks.json")
	if data, err := os.ReadFile(hooksPath); err == nil {
		var config map[string]protocol.NamedHookConfig
		if err := json.Unmarshal(data, &config); err == nil {
			if fence, ok := config["agentfence"]; ok {
				fence.Enabled = false
				config["agentfence"] = fence
				updatedData, _ := json.MarshalIndent(config, "", "  ")
				if err := os.WriteFile(hooksPath, updatedData, 0644); err == nil {
					ui.Pass("Antigravity hooks disabled (.agents/hooks.json: enabled = false)")
					stoppedSomething = true
				}
			}
		}
	}

	fmt.Println()
	if stoppedSomething {
		ui.Warn("AgentFence is now STOPPED.")
		ui.Info("AI agents are running without guardrails or verification.")
		ui.Info("To resume protection at any time, run:")
		ui.Info("  agentfence start")
	} else {
		ui.Info("AgentFence was not active in this repository.")
	}
	fmt.Println()
}

// RunStart re-enables AgentFence hooks.
func RunStart(args []string) {
	fmt.Print(ui.Banner())
	ui.Section("Starting AgentFence Protection")

	cwd, err := os.Getwd()
	if err != nil {
		ui.Fail("Failed to determine current working directory: %v", err)
		os.Exit(1)
	}

	startedSomething := false

	// 1. Enable Antigravity hooks.json
	hooksPath := filepath.Join(cwd, ".agents", "hooks.json")
	if data, err := os.ReadFile(hooksPath); err == nil {
		var config map[string]protocol.NamedHookConfig
		if err := json.Unmarshal(data, &config); err == nil {
			if fence, ok := config["agentfence"]; ok {
				fence.Enabled = true
				config["agentfence"] = fence
				updatedData, _ := json.MarshalIndent(config, "", "  ")
				if err := os.WriteFile(hooksPath, updatedData, 0644); err == nil {
					ui.Pass("Antigravity hooks enabled (.agents/hooks.json: enabled = true)")
					startedSomething = true
				}
			}
		}
	} else {
		// If hooks.json is missing, generate it
		RunInit([]string{})
		return
	}

	fmt.Println()
	if startedSomething {
		ui.Pass("AgentFence is now ACTIVE and protecting this repository.")
	} else {
		ui.Warn("Run 'agentfence init' to configure this project.")
	}
	fmt.Println()
}
