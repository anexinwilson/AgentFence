package verification

import (
	"fmt"
	"strings"

	"github.com/agentfence/agentfence/pkg/api"
)

// FormatAgentStopMessage formats actionable feedback for the AI agent when Stop is blocked.
func FormatAgentStopMessage(rep api.VerificationReport) string {
	if rep.AllPassed {
		return "AgentFence verification passed. All requirements and checks satisfied."
	}

	var sb strings.Builder
	sb.WriteString("AgentFence prevented completion because verification checks failed.\n\n")
	sb.WriteString("Failures:\n")

	for _, res := range rep.Results {
		if !res.Passed {
			firstLine := strings.Split(res.Details, "\n")[0]
			sb.WriteString(fmt.Sprintf("• [%s]: %s\n", strings.ToUpper(res.Name), firstLine))
			if res.ActionableFix != "" {
				sb.WriteString(fmt.Sprintf("  Action needed: %s\n", res.ActionableFix))
			}
		}
	}

	sb.WriteString("\nResolve these issues before claiming the task is complete.")
	return sb.String()
}
