package artifacts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

// Validator checks required output artifacts to prevent fake/mock progress.
type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Name() string {
	return "artifacts"
}

func (v *Validator) Validate(p core.GuardrailsConfig, repoRoot string) api.ValidationResult {
	artifacts := p.Verification.Artifacts
	if len(artifacts) == 0 {
		return api.ValidationResult{
			Name:    v.Name(),
			Passed:  true,
			Details: "No required build artifacts configured (skipped).",
		}
	}

	var missing []string
	var tooSmall []string
	var valid []string

	for _, item := range artifacts {
		target := filepath.Join(repoRoot, filepath.FromSlash(item.Path))
		fi, err := os.Stat(target)
		if err != nil {
			missing = append(missing, item.Path)
			continue
		}

		if item.MinSizeBytes > 0 && fi.Size() < item.MinSizeBytes {
			tooSmall = append(tooSmall, fmt.Sprintf("%s (%d bytes < min %d)", item.Path, fi.Size(), item.MinSizeBytes))
			continue
		}

		valid = append(valid, fmt.Sprintf("%s (%d bytes)", item.Path, fi.Size()))
	}

	allPassed := len(missing) == 0 && len(tooSmall) == 0

	var lines []string
	lines = append(lines, fmt.Sprintf("Artifacts status (%d/%d verified):", len(valid), len(artifacts)))
	for _, v := range valid {
		lines = append(lines, fmt.Sprintf("  ✓ %s", v))
	}
	for _, m := range missing {
		lines = append(lines, fmt.Sprintf("  ✗ %s (does not exist)", m))
	}
	for _, s := range tooSmall {
		lines = append(lines, fmt.Sprintf("  ✗ %s", s))
	}

	var actionable string
	if !allPassed {
		actionable = "Ensure the build, compilation, or bundle process actually ran and generated real output artifacts."
	}

	return api.ValidationResult{
		Name:          v.Name(),
		Passed:        allPassed,
		Details:       strings.Join(lines, "\n"),
		ActionableFix: actionable,
	}
}
