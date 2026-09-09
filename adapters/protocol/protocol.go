package protocol

// ToolCall represents an invoked tool within the Antigravity agent loop.
type ToolCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

// PreToolUseInput matches Antigravity's PreToolUse hook payload on stdin.
type PreToolUseInput struct {
	ToolCall              ToolCall `json:"toolCall"`
	StepIdx               int      `json:"stepIdx"`
	ConversationID        string   `json:"conversationId"`
	WorkspacePaths        []string `json:"workspacePaths"`
	TranscriptPath        string   `json:"transcriptPath"`
	ArtifactDirectoryPath string   `json:"artifactDirectoryPath"`
	ModelName             string   `json:"modelName"`
}

// PreToolUseOutput is sent to stdout for Antigravity's PreToolUse event.
type PreToolUseOutput struct {
	Decision            string   `json:"decision"` // "allow", "deny", "ask", "force_ask"
	Reason              string   `json:"reason,omitempty"`
	PermissionOverrides []string `json:"permissionOverrides,omitempty"`
}

// StopInput matches Antigravity's Stop hook payload on stdin.
type StopInput struct {
	ExecutionNum          int      `json:"executionNum"`
	TerminationReason     string   `json:"terminationReason"`
	Error                 string   `json:"error,omitempty"`
	FullyIdle             bool     `json:"fullyIdle"`
	ConversationID        string   `json:"conversationId"`
	WorkspacePaths        []string `json:"workspacePaths"`
	TranscriptPath        string   `json:"transcriptPath"`
	ArtifactDirectoryPath string   `json:"artifactDirectoryPath"`
	ModelName             string   `json:"modelName"`
}

// StopOutput is sent to stdout for Antigravity's Stop event.
type StopOutput struct {
	Decision string `json:"decision"` // "continue" to force agent to keep working, or "allow"
	Reason   string `json:"reason,omitempty"`
}
