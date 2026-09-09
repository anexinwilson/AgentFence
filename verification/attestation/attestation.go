package attestation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	// SchemaVersion defines the attestation schema revision.
	SchemaVersion = "1.0"
	// LockFileName is the canonical relative path for the attestation file.
	LockFileName = ".agentfence/attestation.json"
)

// Attestation records cryptographic proof that a specific Git tree state passed verification.
type Attestation struct {
	SchemaVersion string    `json:"schema_version"`
	GitTreeSHA    string    `json:"git_tree_sha"`
	WorkingHash   string    `json:"working_hash"`
	Timestamp     time.Time `json:"timestamp"`
	VerifiedBy    string    `json:"verified_by"`
	Passed        bool      `json:"passed"`
	Summary       string    `json:"summary"`
}

// ComputeGitTreeSHA runs 'git write-tree' to obtain the exact SHA of the staged index.
func ComputeGitTreeSHA(repoRoot string) (string, error) {
	cmd := exec.Command("git", "write-tree")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to compute git tree: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ComputeWorkingHash calculates a SHA-256 digest of unstaged changes and status.
// This ensures that modifying files without staging them still invalidates the attestation.
func ComputeWorkingHash(repoRoot string) (string, error) {
	cmdStatus := exec.Command("git", "status", "--porcelain")
	cmdStatus.Dir = repoRoot
	statusOut, err := cmdStatus.Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}

	var filteredStatus []string
	for _, line := range strings.Split(string(statusOut), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.Contains(trimmed, ".agentfence/") {
			continue
		}
		filteredStatus = append(filteredStatus, line)
	}

	cmdDiff := exec.Command("git", "diff")
	cmdDiff.Dir = repoRoot
	diffOut, err := cmdDiff.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}

	h := sha256.New()
	h.Write([]byte(strings.Join(filteredStatus, "\n")))
	h.Write([]byte("---SEPARATOR---"))
	h.Write(diffOut)

	return hex.EncodeToString(h.Sum(nil)), nil
}

// Generate creates and writes the signed attestation file after verification succeeds.
func Generate(repoRoot string, summary string) (*Attestation, error) {
	if repoRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		repoRoot = cwd
	}

	treeSHA, err := ComputeGitTreeSHA(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("git write-tree failed (attestation requires an initialized git repository): %w", err)
	}

	workingHash, err := ComputeWorkingHash(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("working tree hash calculation failed: %w", err)
	}

	att := &Attestation{
		SchemaVersion: SchemaVersion,
		GitTreeSHA:    treeSHA,
		WorkingHash:   workingHash,
		Timestamp:     time.Now().UTC(),
		VerifiedBy:    "agentfence",
		Passed:        true,
		Summary:       summary,
	}

	lockPath := filepath.Join(repoRoot, filepath.FromSlash(LockFileName))
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create attestation directory: %w", err)
	}

	data, err := json.MarshalIndent(att, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal attestation: %w", err)
	}

	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write attestation lockfile: %w", err)
	}

	return att, nil
}

// Verify checks whether the current working tree matches the state recorded in attestation.json.
// Returns (isValid, reason, error).
func Verify(repoRoot string) (bool, string, error) {
	if repoRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return false, "", err
		}
		repoRoot = cwd
	}

	lockPath := filepath.Join(repoRoot, filepath.FromSlash(LockFileName))
	data, err := os.ReadFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "No attestation lockfile found (.agentfence/attestation.json missing). Run 'agentfence verify'.", nil
		}
		return false, "", fmt.Errorf("failed to read attestation lockfile: %w", err)
	}

	var att Attestation
	if err := json.Unmarshal(data, &att); err != nil {
		return false, "Attestation lockfile is corrupt. Run 'agentfence verify'.", nil
	}

	if !att.Passed {
		return false, "Previous verification run failed. Run 'agentfence verify'.", nil
	}

	currentTree, err := ComputeGitTreeSHA(repoRoot)
	if err != nil {
		return false, fmt.Sprintf("Failed to compute git tree: %v", err), err
	}
	if currentTree != att.GitTreeSHA {
		return false, fmt.Sprintf("Working tree modified since last verification (tree mismatch: expected %s, got %s). Run 'agentfence verify'.", att.GitTreeSHA, currentTree), nil
	}

	currentWorking, err := ComputeWorkingHash(repoRoot)
	if err != nil {
		return false, fmt.Sprintf("Failed to compute working copy digest: %v", err), err
	}
	if currentWorking != att.WorkingHash {
		return false, "Unstaged or working tree changes detected after last verification. Run 'agentfence verify'.", nil
	}

	return true, "Attestation valid: working tree matches verified state.", nil
}
