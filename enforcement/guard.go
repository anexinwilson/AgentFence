package enforcement

import (
	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

// Guard defines the interface for PreToolUse interceptors.
type Guard interface {
	Name() string
	Check(toolName string, toolArgs map[string]any, p core.GuardrailsConfig) *api.GuardResult
}
