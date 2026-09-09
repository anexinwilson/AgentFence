package attestation

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAttestationGenerateAndVerify(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "attestation-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize real Git repo to verify true tree SHA
	cmdInit := exec.Command("git", "init")
	cmdInit.Dir = tempDir
	if err := cmdInit.Run(); err != nil {
		t.Fatalf("failed to git init: %v", err)
	}

	testFile := filepath.Join(tempDir, "hello.txt")
	_ = os.WriteFile(testFile, []byte("hello agentfence"), 0644)
	cmdAdd := exec.Command("git", "add", "hello.txt")
	cmdAdd.Dir = tempDir
	_ = cmdAdd.Run()

	// Initially, verify should fail because no attestation exists
	valid, reason, err := Verify(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Errorf("expected verify to return false on missing lockfile, got true")
	}
	if reason == "" {
		t.Errorf("expected non-empty reason")
	}

	// Generate attestation
	att, err := Generate(tempDir, "Tests passed: 4/4")
	if err != nil {
		t.Fatalf("failed to generate attestation: %v", err)
	}
	if att == nil || !att.Passed {
		t.Fatalf("expected attestation to be passed")
	}

	// Verify attestation lockfile exists
	lockPath := filepath.Join(tempDir, filepath.FromSlash(LockFileName))
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("expected lockfile at %s to exist: %v", lockPath, err)
	}

	// Now verify should pass
	valid, reason, err = Verify(tempDir)
	if err != nil {
		t.Fatalf("unexpected error on verify: %v", err)
	}
	if !valid {
		t.Errorf("expected verify to succeed, failed with: %s", reason)
	}
}
