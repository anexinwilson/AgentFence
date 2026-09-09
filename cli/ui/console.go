package ui

import (
	"fmt"
	"strings"

	"github.com/agentfence/agentfence/pkg/api"
)

func Pass(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s✓%s %s\n", Green+Bold, Reset, msg)
}

func Fail(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s✗%s %s\n", Red+Bold, Reset, msg)
}

func Warn(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s!%s %s\n", Yellow+Bold, Reset, msg)
}

func Info(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s•%s %s\n", Cyan, Reset, msg)
}

func Section(title string) {
	fmt.Printf("\n%s%s=== %s ===%s\n", Bold, Cyan, title, Reset)
}

// PrintReport prints the structured verification results to the console.
func PrintReport(rep api.VerificationReport) {
	Section("Verification Results")

	for _, res := range rep.Results {
		if res.Passed {
			Pass("[%s] %s", strings.ToUpper(res.Name), res.Details)
		} else {
			Fail("[%s] %s", strings.ToUpper(res.Name), res.Details)
			if res.ActionableFix != "" {
				fmt.Printf("    %sAction:%s %s\n", Yellow, Reset, res.ActionableFix)
			}
		}
	}

	fmt.Println()
	if rep.AllPassed {
		fmt.Printf("%s%sVERIFICATION PASSED: All guardrails and requirements met.%s\n\n", Green, Bold, Reset)
	} else {
		fmt.Printf("%s%sVERIFICATION FAILED: Fix the issues above before finishing.%s\n\n", Red, Bold, Reset)
	}
}
