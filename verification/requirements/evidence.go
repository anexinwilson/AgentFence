package requirements

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/agentfence/agentfence/common/env"
	"github.com/agentfence/agentfence/core"
)

// EvidenceStatus tracks objective verification of a single requirement.
type EvidenceStatus struct {
	Item      core.RequirementItem
	Satisfied bool
	Reason    string
}

// VerifyEvidence deterministically validates the evidence associated with a requirement.
func VerifyEvidence(item core.RequirementItem, repoRoot string) EvidenceStatus {
	// 1. Check evidence file if specified
	if item.EvidenceFile != "" {
		target := filepath.Join(repoRoot, filepath.FromSlash(item.EvidenceFile))
		fi, err := os.Stat(target)
		if err != nil {
			return EvidenceStatus{
				Item:      item,
				Satisfied: false,
				Reason:    fmt.Sprintf("Evidence file '%s' does not exist.", item.EvidenceFile),
			}
		}

		if fi.Size() == 0 {
			return EvidenceStatus{
				Item:      item,
				Satisfied: false,
				Reason:    fmt.Sprintf("Evidence file '%s' is empty (0 bytes).", item.EvidenceFile),
			}
		}
	}

	// 2. Check evidence command if specified
	if item.EvidenceCommand != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", item.EvidenceCommand)
		} else {
			cmd = exec.CommandContext(ctx, "sh", "-c", item.EvidenceCommand)
		}
		cmd.Dir = repoRoot
		cmd.Env = env.SubprocessEnv()

		out, err := cmd.CombinedOutput()
		if err != nil {
			return EvidenceStatus{
				Item:      item,
				Satisfied: false,
				Reason:    fmt.Sprintf("Evidence command '%s' failed: %v (%s)", item.EvidenceCommand, err, strings.TrimSpace(string(out))),
			}
		}
	}

	evidenceLabel := item.EvidenceFile
	if evidenceLabel == "" {
		evidenceLabel = item.EvidenceCommand
	}
	if evidenceLabel == "" {
		evidenceLabel = "declaration"
	}

	return EvidenceStatus{
		Item:      item,
		Satisfied: true,
		Reason:    fmt.Sprintf("Verified via %s", evidenceLabel),
	}
}
