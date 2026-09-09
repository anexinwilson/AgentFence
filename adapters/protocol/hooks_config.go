package protocol

import "encoding/json"

// HookItem defines a single command hook.
type HookItem struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

// GroupedHook wraps tool matchers with hook items.
type GroupedHook struct {
	Matcher string     `json:"matcher"`
	Hooks   []HookItem `json:"hooks"`
}

// NamedHookConfig defines lifecycle events for a named hook block.
type NamedHookConfig struct {
	Enabled    bool          `json:"enabled"`
	PreToolUse []GroupedHook `json:"PreToolUse,omitempty"`
	Stop       []HookItem    `json:"Stop,omitempty"`
}

// GenerateHooksJSON produces the standard .agents/hooks.json configuration for AgentFence.
func GenerateHooksJSON(binaryCommand string) (string, error) {
	if binaryCommand == "" {
		binaryCommand = "agentfence"
	}

	config := map[string]NamedHookConfig{
		"agentfence": {
			Enabled: true,
			PreToolUse: []GroupedHook{
				{
					Matcher: "run_command|write_to_file|replace_file_content|edit_file|delete_file",
					Hooks: []HookItem{
						{
							Type:    "command",
							Command: binaryCommand + " hook pre-tool",
							Timeout: 15,
						},
					},
				},
			},
			Stop: []HookItem{
				{
					Type:    "command",
					Command: binaryCommand + " hook stop",
					Timeout: 180,
				},
			},
		},
	}

	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
