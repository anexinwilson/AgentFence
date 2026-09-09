package requirements

import (
	"fmt"
	"strings"

	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

// Validator checks requirements checklist for objective evidence.
type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Name() string {
	return "requirements"
}

func (v *Validator) Validate(p core.GuardrailsConfig, repoRoot string) api.ValidationResult {
	checklist := p.Requirements.Checklist
	if len(checklist) == 0 {
		return api.ValidationResult{
			Name:    v.Name(),
			Passed:  true,
			Details: "No requirements defined in checklist.",
		}
	}

	var statuses []EvidenceStatus
	for _, item := range checklist {
		st := VerifyEvidence(item, repoRoot)
		statuses = append(statuses, st)
	}

	satisfiedCount := 0
	for _, s := range statuses {
		if s.Satisfied {
			satisfiedCount++
		}
	}

	totalCount := len(statuses)
	allPassed := satisfiedCount == totalCount

	var lines []string
	lines = append(lines, fmt.Sprintf("Requirements status (%d/%d):", satisfiedCount, totalCount))
	for _, s := range statuses {
		sym := "✓"
		if !s.Satisfied {
			sym = "✗"
		}
		lines = append(lines, fmt.Sprintf("  %s [%s] %s: %s", sym, s.Item.ID, s.Item.Title, s.Reason))
	}

	var actionable string
	if !allPassed {
		var missing []string
		for _, s := range statuses {
			if !s.Satisfied {
				missing = append(missing, s.Item.ID)
			}
		}
		actionable = fmt.Sprintf("Provide required evidence (code files or passing tests) for: %s", strings.Join(missing, ", "))
	}

	return api.ValidationResult{
		Name:          v.Name(),
		Passed:        allPassed,
		Details:       strings.Join(lines, "\n"),
		ActionableFix: actionable,
	}
}
