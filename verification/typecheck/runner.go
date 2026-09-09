package typecheck

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

// CommandRunner validates general build, typecheck, or lint commands.
type CommandRunner struct {
	name   string
	getter func(p core.GuardrailsConfig) *core.VerificationCommand
}

func NewTypecheckRunner() *CommandRunner {
	return &CommandRunner{
		name: "typecheck",
		getter: func(p core.GuardrailsConfig) *core.VerificationCommand {
			return p.Verification.Typecheck
		},
	}
}

func NewBuildRunner() *CommandRunner {
	return &CommandRunner{
		name: "build",
		getter: func(p core.GuardrailsConfig) *core.VerificationCommand {
			return p.Verification.Build
		},
	}
}

func NewLintRunner() *CommandRunner {
	return &CommandRunner{
		name: "lint",
		getter: func(p core.GuardrailsConfig) *core.VerificationCommand {
			return p.Verification.Lint
		},
	}
}

func (c *CommandRunner) Name() string {
	return c.name
}

func (c *CommandRunner) Validate(p core.GuardrailsConfig, repoRoot string) api.ValidationResult {
	cmdCfg := c.getter(p)
	if cmdCfg == nil || strings.TrimSpace(cmdCfg.Command) == "" {
		return api.ValidationResult{
			Name:    c.Name(),
			Passed:  true,
			Details: fmt.Sprintf("No %s command configured (skipped).", c.Name()),
		}
	}

	cmdStr := strings.TrimSpace(cmdCfg.Command)
	timeoutSec := cmdCfg.Timeout
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
			Name:           c.Name(),
			Passed:         false,
			Details:        fmt.Sprintf("%s command timed out after %d seconds: `%s`", strings.ToUpper(c.Name()), timeoutSec, cmdStr),
			ActionableFix:  fmt.Sprintf("Check why %s is hanging or increase timeout.", c.Name()),
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
			Name:           c.Name(),
			Passed:         false,
			Details:        fmt.Sprintf("%s failed (`%s`):\n%s", c.Name(), cmdStr, strings.Join(snippet, "\n")),
			ActionableFix:  fmt.Sprintf("Fix %s errors. Run `%s` locally.", c.Name(), cmdStr),
			Elapsed:        elapsed,
			ExecutionTimeS: elapsedSec,
		}
	}

	return api.ValidationResult{
		Name:           c.Name(),
		Passed:         true,
		Details:        fmt.Sprintf("%s succeeded in %.2fs (`%s`).", strings.ToUpper(c.Name()), elapsedSec, cmdStr),
		Elapsed:        elapsed,
		ExecutionTimeS: elapsedSec,
	}
}
