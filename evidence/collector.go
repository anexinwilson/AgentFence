package evidence

import (
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/verification/artifacts"
	"github.com/agentfence/agentfence/verification/fallbacks"
	"github.com/agentfence/agentfence/verification/placeholders"
	"github.com/agentfence/agentfence/verification/requirements"
	"github.com/agentfence/agentfence/verification/tests"
	"github.com/agentfence/agentfence/verification/typecheck"
)

// Collector gathers objective evidence across requirements, validators, files, and git.
type Collector struct{}

func NewCollector() *Collector {
	return &Collector{}
}

// Collect compiles the full Evidence Report for the repository.
func (c *Collector) Collect(p core.GuardrailsConfig, projectName, repoRoot string) Report {
	rep := Report{
		ProjectName: projectName,
		Timestamp:   time.Now(),
		Verdict:     "VERIFIED",
	}

	allPassed := true

	// 1. Requirements Evidence
	for _, item := range p.Requirements.Checklist {
		st := requirements.VerifyEvidence(item, repoRoot)
		rep.Requirements = append(rep.Requirements, RequirementEvidence{
			ID:        item.ID,
			Title:     item.Title,
			Satisfied: st.Satisfied,
			Proof:     st.Reason,
		})
		if !st.Satisfied {
			allPassed = false
		}
	}

	// 2. Verification Validators
	// Tests
	testRunner := tests.NewRunner()
	testRes := testRunner.Validate(p, repoRoot)
	rep.Verification = append(rep.Verification, VerificationEvidence{
		Name:    "Tests",
		Passed:  testRes.Passed,
		Summary: testRes.Details,
	})
	if !testRes.Passed {
		allPassed = false
	}

	// Placeholders
	placeholderScanner := placeholders.NewScanner()
	placeholderRes := placeholderScanner.Validate(p, repoRoot)
	rep.Verification = append(rep.Verification, VerificationEvidence{
		Name:    "Placeholders",
		Passed:  placeholderRes.Passed,
		Summary: placeholderRes.Details,
	})
	if !placeholderRes.Passed {
		allPassed = false
	}

	// Fallbacks
	fallbackScanner := fallbacks.NewScanner()
	fallbackRes := fallbackScanner.Validate(p, repoRoot)
	rep.Verification = append(rep.Verification, VerificationEvidence{
		Name:    "Fallbacks",
		Passed:  fallbackRes.Passed,
		Summary: fallbackRes.Details,
	})
	if !fallbackRes.Passed {
		allPassed = false
	}

	// Artifacts (if configured)
	if len(p.Verification.Artifacts) > 0 {
		artifactValidator := artifacts.NewValidator()
		artRes := artifactValidator.Validate(p, repoRoot)
		rep.Verification = append(rep.Verification, VerificationEvidence{
			Name:    "Artifacts",
			Passed:  artRes.Passed,
			Summary: artRes.Details,
		})
		if !artRes.Passed {
			allPassed = false
		}
	}

	// Typecheck (if configured)
	if p.Verification.Typecheck != nil && strings.TrimSpace(p.Verification.Typecheck.Command) != "" {
		tcRunner := typecheck.NewTypecheckRunner()
		tcRes := tcRunner.Validate(p, repoRoot)
		rep.Verification = append(rep.Verification, VerificationEvidence{
			Name:    "Typecheck",
			Passed:  tcRes.Passed,
			Summary: tcRes.Details,
		})
		if !tcRes.Passed {
			allPassed = false
		}
	}

	// Build (if configured)
	if p.Verification.Build != nil && strings.TrimSpace(p.Verification.Build.Command) != "" {
		buildRunner := typecheck.NewBuildRunner()
		bRes := buildRunner.Validate(p, repoRoot)
		rep.Verification = append(rep.Verification, VerificationEvidence{
			Name:    "Build",
			Passed:  bRes.Passed,
			Summary: bRes.Details,
		})
		if !bRes.Passed {
			allPassed = false
		}
	}

	// 3. Files changed / created via Git status
	rep.Files = collectChangedFiles(repoRoot)

	// 4. Git Policy Compliance
	rep.Git = collectGitCompliance(p, repoRoot)
	for _, g := range rep.Git {
		if !g.Compliant {
			allPassed = false
		}
	}

	if allPassed {
		rep.Verdict = "VERIFIED"
	} else {
		rep.Verdict = "FAILED"
	}

	return rep
}

func collectChangedFiles(repoRoot string) []FileEvidence {
	var files []FileEvidence
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return files
	}

	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		trimmed := strings.TrimRight(l, "\r\n")
		if len(trimmed) < 4 {
			continue
		}
		statusCodes := trimmed[:2]
		filePath := strings.TrimSpace(trimmed[3:])

		status := "modified"
		if strings.Contains(statusCodes, "?") {
			status = "created"
		} else if strings.Contains(statusCodes, "A") {
			status = "added"
		} else if strings.Contains(statusCodes, "D") {
			status = "deleted"
		}

		files = append(files, FileEvidence{
			Path:   filepath.ToSlash(filePath),
			Status: status,
		})
	}
	return files
}

func collectGitCompliance(p core.GuardrailsConfig, repoRoot string) []GitEvidence {
	var results []GitEvidence

	blocksPush := false
	for _, b := range p.Git.Block {
		if strings.TrimSpace(strings.ToLower(b)) == "push" {
			blocksPush = true
			break
		}
	}

	if blocksPush {
		results = append(results, GitEvidence{
			Rule:      "No remote push",
			Compliant: true,
			Details:   "Direct git push intercepted and prevented by AgentFence",
		})
	}

	cmd := exec.Command("git", "status", "-sb")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err == nil {
		statusStr := string(out)
		if strings.Contains(statusStr, "[ahead") {
			results = append(results, GitEvidence{
				Rule:      "Unpushed local commits",
				Compliant: true,
				Details:   "Commits staged locally; no unauthorized push to remote",
			})
		}
	}

	return results
}
