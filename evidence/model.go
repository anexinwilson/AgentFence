package evidence

import "time"

// RequirementEvidence captures proof for a single requirement.
type RequirementEvidence struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Satisfied bool   `json:"satisfied"`
	Proof     string `json:"proof"`
}

// VerificationEvidence captures proof for a validator (tests, typecheck, build, etc.).
type VerificationEvidence struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Summary string `json:"summary"`
}

// FileEvidence records created or modified files in the session.
type FileEvidence struct {
	Path   string `json:"path"`
	Status string `json:"status"` // "modified", "created", "added", "deleted"
}

// GitEvidence records Git compliance status.
type GitEvidence struct {
	Rule      string `json:"rule"`
	Compliant bool   `json:"compliant"`
	Details   string `json:"details"`
}

// Report holds the complete objective proof bundle for a task.
type Report struct {
	ProjectName  string                 `json:"projectName"`
	Timestamp    time.Time              `json:"timestamp"`
	Requirements []RequirementEvidence  `json:"requirements"`
	Verification []VerificationEvidence `json:"verification"`
	Files        []FileEvidence         `json:"files"`
	Git          []GitEvidence          `json:"git"`
	Verdict      string                 `json:"verdict"` // "VERIFIED" or "FAILED"
	SummaryText  string                 `json:"summaryText"`
}
