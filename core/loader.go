package core

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CandidateNames are filenames AgentFence searches for when locating configuration.
var CandidateNames = []string{
	"agentfence.yaml",
	"agentfence.yml",
	"agentguard.yaml",
	"agentguard.yml",
}

// FindConfigFile walks upwards from startDir looking for an AgentFence config file.
func FindConfigFile(startDir string) (string, error) {
	if startDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		startDir = cwd
	}

	current, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		for _, name := range CandidateNames {
			candidate := filepath.Join(current, name)
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
				return candidate, nil
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", os.ErrNotExist
}

// Load loads and parses the policy configuration from filePath or nearest ancestor directory.
func Load(filePath string) (Config, error) {
	if filePath == "" {
		found, err := FindConfigFile("")
		if err != nil {
			return DefaultConfig(""), nil
		}
		filePath = found
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read policy file %s: %w", filePath, err)
	}

	cfg := DefaultConfig("")
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid YAML in %s: %w", filePath, err)
	}

	return cfg, nil
}
