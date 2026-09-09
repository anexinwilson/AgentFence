package verification

import (
	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

// Validator defines the interface for deterministic verification checks.
type Validator interface {
	Name() string
	Validate(p core.GuardrailsConfig, repoRoot string) api.ValidationResult
}
