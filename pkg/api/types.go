package api

import "time"

// Decision represents the enforcement action returned to the agent loop.
type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
	DecisionAsk   Decision = "ask"
)

// GuardResult captures the result of a PreToolUse inspection.
type GuardResult struct {
	Decision  Decision `json:"decision"`
	Reason    string   `json:"reason,omitempty"`
	GuardName string   `json:"guardName,omitempty"`
	RuleName  string   `json:"ruleName,omitempty"`
}

func (g GuardResult) IsBlocked() bool {
	return g.Decision == DecisionDeny || g.Decision == DecisionAsk
}

// ValidationResult represents an individual validator check (e.g. tests, lint, placeholders).
type ValidationResult struct {
	Name           string        `json:"name"`
	Passed         bool          `json:"passed"`
	Details        string        `json:"details"`
	ActionableFix  string        `json:"actionableFix,omitempty"`
	Elapsed        time.Duration `json:"elapsed"`
	ExecutionTimeS float64       `json:"executionTimeSeconds"`
}

// StatusSymbol returns a visual indicator for terminal reports.
func (v ValidationResult) StatusSymbol() string {
	if v.Passed {
		return "✓"
	}
	return "✗"
}

// VerificationReport contains the aggregated results from all active validators.
type VerificationReport struct {
	AllPassed bool               `json:"allPassed"`
	Results   []ValidationResult `json:"results"`
}
