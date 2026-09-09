package core

import "fmt"

// DefaultConfig returns a standard baseline configuration.
func DefaultConfig(projectName string) Config {
	if projectName == "" {
		projectName = "my-project"
	}
	return Config{
		Project: ProjectConfig{
			Name: projectName,
		},
		Guardrails: GuardrailsConfig{
			Git: GitPolicy{
				Block:         []string{"push", "commit", "reset --hard", "clean -f", "rebase"},
				AllowReadOnly: true,
				Action:        "deny",
			},
			Files: FilesPolicy{
				Protected: []string{
					".env*",
					"secrets/**",
					"**/*.pem",
					"**/*.key",
					"credentials.json",
				},
				Action: "deny",
			},
			Commands: CommandsPolicy{
				Block: []string{
					"rm -rf *",
					"rm -rf /",
					"curl * | bash",
					"curl * | sh",
					"wget * | bash",
					"wget * | sh",
					":(){ :|:& };:",
				},
				Action: "deny",
			},
			Placeholders: PlaceholdersPolicy{
				Block: true,
				ScanExtensions: []string{
					".go",
					".py",
					".ts",
					".js",
					".rs",
					".java",
				},
				SuspiciousPatterns: []string{
					"TODO",
					"FIXME",
					"XXX",
					"PLACEHOLDER",
					"MOCK_IMPLEMENTATION",
				},
				Action: "fail",
			},
			Fallbacks: FallbacksPolicy{
				Block: true,
				ScanExtensions: []string{
					".go",
					".py",
					".ts",
					".js",
					".rs",
				},
				Action: "fail",
			},
			Verification: VerificationPolicy{
				Tests: &VerificationCommand{
					Command:  "go test ./...",
					Required: true,
					Timeout:  60,
				},
			},
			Requirements: RequirementsPolicy{
				Checklist: []RequirementItem{},
			},
		},
	}
}

// GenerateDefaultYAML returns a well-commented agentfence.yaml configuration tailored to the project ecosystem.
func GenerateDefaultYAML(projectName string, repoRoot string) string {
	eco := DetectEcosystem(repoRoot)
	if projectName == "" {
		projectName = eco.Name
	}
	if projectName == "" {
		projectName = "my-project"
	}

	testCmd := eco.TestCommand
	if testCmd == "" {
		testCmd = "echo 'No tests configured'"
	}

	typecheckBlock := ""
	if eco.TypecheckCommand != "" {
		typecheckBlock = fmt.Sprintf(`    typecheck:
      command: "%s"
      required: true
      timeout: 60
`, eco.TypecheckCommand)
	}

	lintBlock := ""
	if eco.LintCommand != "" {
		lintBlock = fmt.Sprintf(`    lint:
      command: "%s"
      required: true
      timeout: 60
`, eco.LintCommand)
	}

	buildBlock := ""
	if eco.BuildCommand != "" {
		buildBlock = fmt.Sprintf(`    build:
      command: "%s"
      required: true
      timeout: 120
`, eco.BuildCommand)
	}

	return fmt.Sprintf(`# AgentFence Policy Configuration
# Documentation: https://github.com/agentfence/agentfence

project:
  name: %s

guardrails:
  # Prohibited git operations for AI coding agents
  git:
    block:
      - push
      - commit
      - reset --hard
      - clean -f
      - rebase
    allow_readonly: true
    action: deny

  # Files protected from unintended agent modifications
  files:
    protected:
      - ".env*"
      - "secrets/**"
      - "**/*.pem"
      - "**/*.key"
      - "credentials.json"
    action: deny

  # Destructive shell command patterns
  commands:
    block:
      - "rm -rf *"
      - "rm -rf /"
      - "curl * | bash"
      - "curl * | sh"
      - "wget * | bash"
      - "wget * | sh"
    action: deny

  # Prevent placeholder, TODO-only, or fake implementations
  placeholders:
    block: true
    scan_extensions:
      - .go
      - .py
      - .ts
      - .tsx
      - .js
      - .rs
    action: fail

  # Prevent silent fallbacks, swallowed exceptions, and dual-system technical debt
  fallbacks:
    block: true
    scan_extensions:
      - .go
      - .py
      - .ts
      - .tsx
      - .js
      - .rs
    action: fail

  # Deterministic verification required before agent completion
  verification:
    tests:
      command: "%s"
      required: true
      timeout: 60
%s%s%s
  # Requirements checklist with verifiable evidence
  requirements:
    checklist: []
    # Add requirements to enforce verifiable delivery:
    # - id: "R-01"
    #   title: "Core feature implementation"
    #   evidence_file: "path/to/evidence.ts"
    #   required: true
`, projectName, testCmd, typecheckBlock, lintBlock, buildBlock)
}
