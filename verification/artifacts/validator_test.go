package artifacts

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentfence/agentfence/core"
)

func TestArtifactsValidator(t *testing.T) {
	tempDir := t.TempDir()
	bundleFile := filepath.Join(tempDir, "bundle.js")
	_ = os.WriteFile(bundleFile, []byte("console.log('hello world');"), 0644)

	validator := NewValidator()

	// 1. Success case: existing file with valid size
	cfg := core.GuardrailsConfig{
		Verification: core.VerificationPolicy{
			Artifacts: []core.ArtifactItem{
				{
					Path:         "bundle.js",
					MinSizeBytes: 10,
					Required:     true,
				},
			},
		},
	}
	res := validator.Validate(cfg, tempDir)
	if !res.Passed {
		t.Errorf("Expected artifact validation to pass, got failure: %s", res.Details)
	}

	// 2. Failure case: missing artifact
	cfgMissing := core.GuardrailsConfig{
		Verification: core.VerificationPolicy{
			Artifacts: []core.ArtifactItem{
				{
					Path:         "missing.bin",
					MinSizeBytes: 10,
					Required:     true,
				},
			},
		},
	}
	resMissing := validator.Validate(cfgMissing, tempDir)
	if resMissing.Passed {
		t.Errorf("Expected artifact validation to fail for missing file")
	}
}
