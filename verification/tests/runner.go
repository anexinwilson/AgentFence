package tests

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/agentfence/agentfence/common/env"
	"github.com/agentfence/agentfence/core"
	"github.com/agentfence/agentfence/pkg/api"
)

// Runner executes test suites and verifies exit status.
type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Name() string {
	return "tests"
}

func (r *Runner) Validate(p core.GuardrailsConfig, repoRoot string) api.ValidationResult {
	testCfg := p.Verification.Tests
	if testCfg == nil || strings.TrimSpace(testCfg.Command) == "" {
		return api.ValidationResult{
			Name:    r.Name(),
			Passed:  true,
			Details: "No test command configured (skipped).",
		}
	}

	cmdStr := strings.TrimSpace(testCfg.Command)
	timeoutSec := testCfg.Timeout
	if timeoutSec <= 0 {
		timeoutSec = 60
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	start := time.Now()
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", cmdStr)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}

	if repoRoot != "" {
		cmd.Dir = repoRoot
	}
	cmd.Env = env.SubprocessEnv()

	outBytes, err := cmd.CombinedOutput()
	elapsed := time.Since(start)
	elapsedSec := elapsed.Seconds()

	if ctx.Err() == context.DeadlineExceeded {
		return api.ValidationResult{
			Name:           r.Name(),
			Passed:         false,
			Details:        fmt.Sprintf("Test command timed out after %d seconds: `%s`", timeoutSec, cmdStr),
			ActionableFix:  "Investigate slow/hanging tests or increase timeout in agentfence.yaml.",
			Elapsed:        elapsed,
			ExecutionTimeS: elapsedSec,
		}
	}

	output := strings.TrimSpace(string(outBytes))

	if err != nil {
		lines := strings.Split(output, "\n")
		var snippet []string
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" {
				snippet = append(snippet, l)
			}
		}
		if len(snippet) > 10 {
			snippet = snippet[len(snippet)-10:]
		}

		return api.ValidationResult{
			Name:           r.Name(),
			Passed:         false,
			Details:        fmt.Sprintf("Test suite failed (`%s`):\n%s", cmdStr, strings.Join(snippet, "\n")),
			ActionableFix:  fmt.Sprintf("Fix failing test cases. Run `%s` locally to inspect full details.", cmdStr),
			Elapsed:        elapsed,
			ExecutionTimeS: elapsedSec,
		}
	}

	return api.ValidationResult{
		Name:           r.Name(),
		Passed:         true,
		Details:        fmt.Sprintf("All tests passed in %.2fs (`%s`).", elapsedSec, cmdStr),
		Elapsed:        elapsed,
		ExecutionTimeS: elapsedSec,
	}
}
