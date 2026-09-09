package verification

import (
	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
	"github.com/agentfence/agentfence/verification/artifacts"
	"github.com/agentfence/agentfence/verification/codecheck"
	"github.com/agentfence/agentfence/verification/fallbacks"
	"github.com/agentfence/agentfence/verification/placeholders"
	"github.com/agentfence/agentfence/verification/requirements"
	"github.com/agentfence/agentfence/verification/tests"
	"github.com/agentfence/agentfence/verification/typecheck"
)

// Engine executes all configured verification checks.
type Engine struct {
	validators []Validator
}

// NewEngine creates a verification engine with the standard validators.
func NewEngine() *Engine {
	return &Engine{
		validators: []Validator{
			requirements.NewValidator(),
			artifacts.NewValidator(),
			codecheck.NewValidator(),
			placeholders.NewScanner(),
			fallbacks.NewScanner(),
			tests.NewRunner(),
			typecheck.NewTypecheckRunner(),
			typecheck.NewBuildRunner(),
			typecheck.NewLintRunner(),
		},
	}
}

// Run executes all validators and returns an aggregated VerificationReport.
func (e *Engine) Run(p core.GuardrailsConfig, repoRoot string) api.VerificationReport {
	var results []api.ValidationResult
	allPassed := true

	for _, v := range e.validators {
		res := v.Validate(p, repoRoot)
		results = append(results, res)
		if !res.Passed {
			allPassed = false
		}
	}

	return api.VerificationReport{
		AllPassed: allPassed,
		Results:   results,
	}
}
