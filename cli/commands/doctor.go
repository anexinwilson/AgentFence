package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/agentfence/agentfence/cli/ui"
	"github.com/agentfence/agentfence/core"
)

// RunDoctor performs a diagnostic audit of AgentFence, detecting global toolchain and project-level status.
func RunDoctor(args []string) {
	fmt.Print(ui.Banner())

	cwd, _ := os.Getwd()
	hasIssues := false
	isProjectConfigured := false

	// Section 1: Global Toolchain
	ui.Section("Global Toolchain")
	if _, err := exec.LookPath("git"); err == nil {
		ui.Pass("Git CLI detected in PATH")
	} else {
		ui.Warn("Git CLI not found in PATH")
	}

	gitDir := filepath.Join(cwd, ".git")
	if fi, err := os.Stat(gitDir); err == nil && fi.IsDir() {
		ui.Pass("Git repository initialized")
	} else {
		ui.Warn("Current directory is not a Git repository root")
	}

	// Section 2: Project Detection & Status
	ui.Section("Project Integration Status")
	eco := core.DetectEcosystem(cwd)
	ui.Info("Detected directory: %s", filepath.Base(cwd))
	ui.Info("Tech stack: %s", eco.Summary())

	configPath, err := core.FindConfigFile(cwd)
	if err == nil {
		isProjectConfigured = true
		ui.Pass("Configuration file: %s", filepath.Base(configPath))
		cfg, err := core.Load(configPath)
		if err == nil {
			ui.Pass("Policy loaded (project name: %s)", cfg.Project.Name)
		} else {
			ui.Fail("Configuration YAML error: %v", err)
			hasIssues = true
		}
	} else {
		ui.Warn("Policy configuration: NOT FOUND (agentfence.yaml missing)")
	}

	// Check Antigravity hooks.json
	hooksConfigured := false
	hooksPath := filepath.Join(cwd, ".agents", "hooks.json")
	if data, err := os.ReadFile(hooksPath); err == nil {
		var dummy map[string]any
		if err := json.Unmarshal(data, &dummy); err == nil {
			hooksConfigured = true
			ui.Pass("Antigravity hooks: ACTIVE (.agents/hooks.json)")
		} else {
			ui.Fail("Antigravity hooks contain invalid JSON: %v", err)
			hasIssues = true
		}
	} else {
		ui.Warn("Antigravity hooks: NOT CONFIGURED (.agents/hooks.json missing)")
	}

	// Check Antigravity rules
	rulesPath := filepath.Join(cwd, ".agents", "rules", "agentfence.md")
	if _, err := os.Stat(rulesPath); err == nil {
		ui.Pass("Agent rules: INSTALLED (.agents/rules/agentfence.md)")
	} else {
		ui.Warn("Agent rules: NOT FOUND (.agents/rules/agentfence.md missing)")
	}

	// Final Summary
	fmt.Println()
	if hasIssues {
		ui.Fail("Doctor detected configuration problems. Please review errors above.")
		os.Exit(1)
	} else if !isProjectConfigured || !hooksConfigured {
		ui.Warn("PROJECT STATUS: AgentFence is installed globally, but NOT fully active in this project.")
		fmt.Printf("  -> Run 'agentfence init' to configure guardrails and hooks for %s (%s).\n\n", eco.Name, eco.Summary())
	} else {
		ui.Pass("PROJECT STATUS: Fully configured and protected by AgentFence.")
	}
	fmt.Println()
}
