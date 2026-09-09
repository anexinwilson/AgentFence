package core

// GitPolicy controls allowed and prohibited Git operations.
type GitPolicy struct {
	Block         []string `yaml:"block"`
	AllowReadOnly bool     `yaml:"allow_readonly"`
	Action        string   `yaml:"action"`
}

// FilesPolicy defines protected file patterns that agents cannot mutate.
type FilesPolicy struct {
	Protected []string `yaml:"protected"`
	Action    string   `yaml:"action"`
}

// CommandsPolicy defines shell command patterns that agents cannot execute.
type CommandsPolicy struct {
	Block  []string `yaml:"block"`
	Action string   `yaml:"action"`
}

// PlaceholdersPolicy controls detection of incomplete or stub implementations.
type PlaceholdersPolicy struct {
	Block              bool     `yaml:"block"`
	ScanExtensions     []string `yaml:"scan_extensions"`
	SuspiciousPatterns []string `yaml:"suspicious_patterns"`
	Action             string   `yaml:"action"`
}

// FallbacksPolicy controls detection of silent fallbacks and swallowed exceptions.
type FallbacksPolicy struct {
	Block          bool     `yaml:"block"`
	ScanExtensions []string `yaml:"scan_extensions"`
	Action         string   `yaml:"action"`
}

// ArtifactItem represents a required output artifact that must exist and be valid.
type ArtifactItem struct {
	Path         string `yaml:"path"`
	MinSizeBytes int64  `yaml:"min_size_bytes"`
	Required     bool   `yaml:"required"`
}

// VerificationCommand configures an external verification process (e.g. tests, build).
type VerificationCommand struct {
	Command  string `yaml:"command"`
	Required bool   `yaml:"required"`
	Timeout  int    `yaml:"timeout"`
}

// VerificationPolicy specifies deterministic verification steps.
type VerificationPolicy struct {
	Tests     *VerificationCommand `yaml:"tests"`
	Typecheck *VerificationCommand `yaml:"typecheck"`
	Lint      *VerificationCommand `yaml:"lint"`
	Build     *VerificationCommand `yaml:"build"`
	Artifacts []ArtifactItem       `yaml:"artifacts"`
}

// RequirementItem represents a user-specified requirement with objective evidence.
type RequirementItem struct {
	ID              string `yaml:"id"`
	Title           string `yaml:"title"`
	EvidenceFile    string `yaml:"evidence_file"`
	EvidenceCommand string `yaml:"evidence_command"`
	Required        bool   `yaml:"required"`
}

// RequirementsPolicy holds the checklist of requirements.
type RequirementsPolicy struct {
	Checklist []RequirementItem `yaml:"checklist"`
}

// GuardrailsConfig wraps all guardrail settings.
type GuardrailsConfig struct {
	Git          GitPolicy          `yaml:"git"`
	Files        FilesPolicy        `yaml:"files"`
	Commands     CommandsPolicy     `yaml:"commands"`
	Placeholders PlaceholdersPolicy `yaml:"placeholders"`
	Fallbacks    FallbacksPolicy    `yaml:"fallbacks"`
	Verification VerificationPolicy `yaml:"verification"`
	Requirements RequirementsPolicy `yaml:"requirements"`
}

// ProjectConfig identifies the target repository.
type ProjectConfig struct {
	Name string `yaml:"name"`
}

// Config represents the complete agentfence.yaml configuration.
type Config struct {
	Project    ProjectConfig    `yaml:"project"`
	Guardrails GuardrailsConfig `yaml:"guardrails"`
}
