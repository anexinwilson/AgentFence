package enforcement

import (
	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/enforcement/commands"
	"github.com/agentfence/agentfence/enforcement/files"
	"github.com/agentfence/agentfence/enforcement/git"
	"github.com/agentfence/agentfence/pkg/api"
)

// Router coordinates multiple guards and determines the final enforcement action.
type Router struct {
	guards []Guard
}

// NewRouter instantiates the standard set of PreToolUse guards.
func NewRouter() *Router {
	return &Router{
		guards: []Guard{
			git.NewGuard(),
			commands.NewGuard(),
			files.NewGuard(),
		},
	}
}

// Evaluate runs all guards against the tool invocation.
func (r *Router) Evaluate(toolName string, toolArgs map[string]any, p core.GuardrailsConfig) api.GuardResult {
	for _, guard := range r.guards {
		res := guard.Check(toolName, toolArgs, p)
		if res != nil && res.IsBlocked() {
			return *res
		}
	}
	return api.GuardResult{
		Decision: api.DecisionAllow,
	}
}
