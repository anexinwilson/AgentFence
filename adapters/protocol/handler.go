package protocol

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/enforcement"
	"github.com/agentfence/agentfence/evidence"
	"github.com/agentfence/agentfence/verification"
	"github.com/agentfence/agentfence/verification/attestation"
)

// HandlePreTool processes an Antigravity PreToolUse hook payload from stdin and writes decision to stdout.
func HandlePreTool(r io.Reader, w io.Writer, p core.GuardrailsConfig) error {
	var input PreToolUseInput
	if err := json.NewDecoder(r).Decode(&input); err != nil {
		out := PreToolUseOutput{
			Decision: "allow",
			Reason:   "AgentFence PreToolUse: empty or unparseable input",
		}
		return json.NewEncoder(w).Encode(out)
	}

	// Dynamically load the project's config from the workspace path if provided
	if len(input.WorkspacePaths) > 0 && input.WorkspacePaths[0] != "" {
		repoRoot := input.WorkspacePaths[0]
		if foundConfig, err := core.FindConfigFile(repoRoot); err == nil {
			if loadedCfg, err := core.Load(foundConfig); err == nil {
				p = loadedCfg.Guardrails
			}
		}
	}

	router := enforcement.NewRouter()
	result := router.Evaluate(input.ToolCall.Name, input.ToolCall.Args, p)

	out := PreToolUseOutput{
		Decision: string(result.Decision),
		Reason:   result.Reason,
	}

	return json.NewEncoder(w).Encode(out)
}

// HandleStop processes an Antigravity Stop hook payload from stdin and writes decision to stdout.
func HandleStop(r io.Reader, w io.Writer, p core.GuardrailsConfig, repoRoot string) error {
	var input StopInput
	_ = json.NewDecoder(r).Decode(&input)

	// Prioritize workspace path provided by Antigravity over fallback working directory
	projectName := filepath.Base(repoRoot)
	if len(input.WorkspacePaths) > 0 && input.WorkspacePaths[0] != "" {
		repoRoot = input.WorkspacePaths[0]
		projectName = filepath.Base(repoRoot)
		if foundConfig, err := core.FindConfigFile(repoRoot); err == nil {
			if loadedCfg, err := core.Load(foundConfig); err == nil {
				p = loadedCfg.Guardrails
				if loadedCfg.Project.Name != "" {
					projectName = loadedCfg.Project.Name
				}
			}
		}
	}

	engine := verification.NewEngine()
	report := engine.Run(p, repoRoot)

	var out StopOutput
	if report.AllPassed {
		// Automatically compile objective evidence upon passing all verification checks
		collector := evidence.NewCollector()
		evReport := collector.Collect(p, projectName, repoRoot)
		mdReport := evidence.FormatEvidenceMarkdown(evReport)

		// 1. Write permanent evidence to .agents/evidence.md in the repository
		agentsDir := filepath.Join(repoRoot, ".agents")
		if fi, err := os.Stat(agentsDir); err == nil && fi.IsDir() {
			_ = os.WriteFile(filepath.Join(agentsDir, "evidence.md"), []byte(mdReport), 0644)
		}

		// 2. Seamlessly emit evidence as an active Antigravity artifact in the chat UI
		if input.ArtifactDirectoryPath != "" {
			if fi, err := os.Stat(input.ArtifactDirectoryPath); err == nil && fi.IsDir() {
				_ = os.WriteFile(filepath.Join(input.ArtifactDirectoryPath, "agentfence_evidence.md"), []byte(mdReport), 0644)
			}
		}

		// 3. Cryptographically lock tree in .agentfence/attestation.json
		_, _ = attestation.Generate(repoRoot, "Stop verification passed: all checks green")

		out = StopOutput{
			Decision: "allow",
			Reason:   "AgentFence verification passed. Objective evidence verified and saved.",
		}
	} else {
		out = StopOutput{
			Decision: "continue",
			Reason:   verification.FormatAgentStopMessage(report),
		}
	}

	return json.NewEncoder(w).Encode(out)
}
