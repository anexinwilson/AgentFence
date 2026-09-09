package protocol

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/agentfence/agentfence/core"
)

func TestHandlePreTool(t *testing.T) {
	p := core.GuardrailsConfig{
		Git: core.GitPolicy{
			Block:  []string{"push"},
			Action: "deny",
		},
	}

	input := PreToolUseInput{
		ToolCall: ToolCall{
			Name: "run_command",
			Args: map[string]any{"CommandLine": "git push origin main"},
		},
	}
	inBytes, _ := json.Marshal(input)
	var outBuf bytes.Buffer

	if err := HandlePreTool(bytes.NewReader(inBytes), &outBuf, p); err != nil {
		t.Fatalf("HandlePreTool error: %v", err)
	}

	var output PreToolUseOutput
	if err := json.Unmarshal(outBuf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if output.Decision != "deny" {
		t.Errorf("Expected decision 'deny', got '%s'", output.Decision)
	}

	inputAllowed := PreToolUseInput{
		ToolCall: ToolCall{
			Name: "run_command",
			Args: map[string]any{"CommandLine": "git status"},
		},
	}
	inBytesAllowed, _ := json.Marshal(inputAllowed)
	outBuf.Reset()

	if err := HandlePreTool(bytes.NewReader(inBytesAllowed), &outBuf, p); err != nil {
		t.Fatalf("HandlePreTool error: %v", err)
	}

	if err := json.Unmarshal(outBuf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if output.Decision != "allow" {
		t.Errorf("Expected decision 'allow', got '%s'", output.Decision)
	}
}

func TestHandleStop(t *testing.T) {
	p := core.GuardrailsConfig{
		Requirements: core.RequirementsPolicy{
			Checklist: []core.RequirementItem{
				{
					ID:           "R-01",
					Title:        "Missing evidence file",
					EvidenceFile: "nonexistent_file_xyz.go",
				},
			},
		},
	}

	var outBuf bytes.Buffer
	if err := HandleStop(bytes.NewReader([]byte("{}")), &outBuf, p, t.TempDir()); err != nil {
		t.Fatalf("HandleStop error: %v", err)
	}

	var output StopOutput
	if err := json.Unmarshal(outBuf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse Stop output JSON: %v", err)
	}

	if output.Decision != "continue" {
		t.Errorf("Expected Stop decision 'continue' due to failed requirement, got '%s'", output.Decision)
	}
}
