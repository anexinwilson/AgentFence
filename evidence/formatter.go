package evidence

import (
	"fmt"
	"strings"

	"github.com/agentfence/agentfence/cli/ui"
)

// FormatEvidenceMarkdown produces clean GitHub-flavored markdown evidence block.
func FormatEvidenceMarkdown(rep Report) string {
	var sb strings.Builder

	sb.WriteString("### Requirements\n")
	sb.WriteString("```text\n")
	if len(rep.Requirements) == 0 {
		sb.WriteString("No explicit requirements defined.\n")
	} else {
		for _, r := range rep.Requirements {
			sym := "✓"
			if !r.Satisfied {
				sym = "✗"
			}
			sb.WriteString(fmt.Sprintf("%s [%s] %s\n    %s\n", sym, r.ID, r.Title, r.Proof))
		}
	}
	sb.WriteString("```\n\n")

	sb.WriteString("### Verification\n")
	sb.WriteString("```text\n")
	for _, v := range rep.Verification {
		sym := "✓"
		if !v.Passed {
			sym = "✗"
		}
		firstLine := strings.Split(v.Summary, "\n")[0]
		sb.WriteString(fmt.Sprintf("%s %s: %s\n", sym, v.Name, firstLine))
	}
	sb.WriteString("```\n\n")

	sb.WriteString("### Files (Session Changes)\n")
	sb.WriteString("```text\n")
	if len(rep.Files) == 0 {
		sb.WriteString("No modified or created files detected.\n")
	} else {
		for _, f := range rep.Files {
			sb.WriteString(fmt.Sprintf("✓ %s (%s)\n", f.Path, f.Status))
		}
	}
	sb.WriteString("```\n\n")

	sb.WriteString("### Git Compliance\n")
	sb.WriteString("```text\n")
	if len(rep.Git) == 0 {
		sb.WriteString("✓ Git guardrails satisfied\n")
	} else {
		for _, g := range rep.Git {
			sym := "✓"
			if !g.Compliant {
				sym = "✗"
			}
			sb.WriteString(fmt.Sprintf("%s %s (%s)\n", sym, g.Rule, g.Details))
		}
	}
	sb.WriteString("```\n\n")

	satisfiedReqs := 0
	for _, r := range rep.Requirements {
		if r.Satisfied {
			satisfiedReqs++
		}
	}

	sb.WriteString("### Evidence Summary\n")
	sb.WriteString("```text\n")
	sb.WriteString(fmt.Sprintf("Requirements: %d/%d satisfied\n", satisfiedReqs, len(rep.Requirements)))
	for _, v := range rep.Verification {
		status := "passed"
		if !v.Passed {
			status = "FAILED"
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", v.Name, status))
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("VERDICT: %s\n", rep.Verdict))
	sb.WriteString("```\n")

	return sb.String()
}

// PrintEvidenceConsole prints the evidence block to stdout with clean terminal formatting.
func PrintEvidenceConsole(rep Report) {
	ui.Section("Requirements Proof")
	for _, r := range rep.Requirements {
		if r.Satisfied {
			ui.Pass("[%s] %s", r.ID, r.Title)
			fmt.Printf("    %sProof:%s %s\n", ui.Gray, ui.Reset, r.Proof)
		} else {
			ui.Fail("[%s] %s", r.ID, r.Title)
			fmt.Printf("    %sReason:%s %s\n", ui.Red, ui.Reset, r.Proof)
		}
	}

	ui.Section("Verification")
	for _, v := range rep.Verification {
		firstLine := strings.Split(v.Summary, "\n")[0]
		if v.Passed {
			ui.Pass("%s: %s", v.Name, firstLine)
		} else {
			ui.Fail("%s: %s", v.Name, firstLine)
		}
	}

	ui.Section("Files")
	if len(rep.Files) == 0 {
		fmt.Printf("  %s(No modified files)%s\n", ui.Gray, ui.Reset)
	} else {
		for _, f := range rep.Files {
			fmt.Printf("  %s✓%s %s %s(%s)%s\n", ui.Green, ui.Reset, f.Path, ui.Gray, f.Status, ui.Reset)
		}
	}

	ui.Section("Git Compliance")
	for _, g := range rep.Git {
		if g.Compliant {
			ui.Pass("%s: %s", g.Rule, g.Details)
		} else {
			ui.Fail("%s: %s", g.Rule, g.Details)
		}
	}

	satisfiedReqs := 0
	for _, r := range rep.Requirements {
		if r.Satisfied {
			satisfiedReqs++
		}
	}

	ui.Section("Evidence Summary")
	fmt.Printf("  Requirements: %d/%d satisfied\n", satisfiedReqs, len(rep.Requirements))
	for _, v := range rep.Verification {
		status := ui.Green + "passed" + ui.Reset
		if !v.Passed {
			status = ui.Red + "FAILED" + ui.Reset
		}
		fmt.Printf("  %s: %s\n", v.Name, status)
	}

	fmt.Println()
	if rep.Verdict == "VERIFIED" {
		fmt.Printf("  %s%sVERDICT: VERIFIED%s\n\n", ui.Green, ui.Bold, ui.Reset)
	} else {
		fmt.Printf("  %s%sVERDICT: FAILED (Unsatisfied Requirements or Failing Verification)%s\n\n", ui.Red, ui.Bold, ui.Reset)
	}
}
