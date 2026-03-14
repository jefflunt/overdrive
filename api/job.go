package api

import (
	"strings"
	"time"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// GenerateJobID generates a short, sortable, alphanumeric ID based on current time
func GenerateJobID() string {
	n := time.Now().UnixNano()
	if n == 0 {
		return "0"
	}

	var b strings.Builder
	for n > 0 {
		b.WriteByte(base62Chars[n%62])
		n /= 62
	}

	// Reverse to keep chronological order sortable
	s := b.String()
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// JobRequest represents the JSON payload for submitting a job
type JobRequest struct {
	Project      string `json:"project" yaml:"project"`
	RepoURL      string `json:"repo_url" yaml:"repo_url"`
	BranchParent string `json:"branch_parent" yaml:"branch_parent"`
	CommitMsg    string `json:"commit_msg" yaml:"commit_msg"`
	Prompt       string `json:"prompt" yaml:"prompt"`
	Model        string `json:"model" yaml:"model"`
	ReferenceID  string `json:"reference_id,omitempty" yaml:"reference_id,omitempty"`
}

// Job represents the full job state stored in YAML
type Job struct {
	ID            string     `json:"id" yaml:"id"`
	Project       string     `json:"project" yaml:"project"`
	Status        string     `json:"status" yaml:"status"` // pending, working, done, crash, no-op, timeout, stopped, undone, cancelled
	Repo          string     `json:"repo" yaml:"repo"`
	CreatedAt     time.Time  `json:"created_at" yaml:"created_at"`
	StartedAt     *time.Time `json:"started_at,omitempty" yaml:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty" yaml:"completed_at,omitempty"`
	Request       JobRequest `json:"request" yaml:"request"`
	RelatedCommit string     `json:"related_commit,omitempty" yaml:"related_commit,omitempty"`
	SubStatus     string     `json:"sub_status,omitempty" yaml:"sub_status,omitempty"`
	TestStatus    string     `json:"test_status,omitempty" yaml:"test_status,omitempty"` // passed, failed, error
	TestOutput    string     `json:"test_output,omitempty" yaml:"test_output,omitempty"`
	ExitCode      *int       `json:"exit_code,omitempty" yaml:"exit_code,omitempty"`
	Error         string     `json:"error,omitempty" yaml:"error,omitempty"`
	ReferenceID   string     `json:"reference_id,omitempty" yaml:"reference_id,omitempty"`
}

// IsTerminal returns true if the job is in a final state
func (j Job) IsTerminal() bool {
	return j.Status == "done" || j.Status == "crash" || j.Status == "no-op" || j.Status == "timeout" || j.Status == "stopped" || j.Status == "undone" || j.Status == "cancelled"
}

// CleanPromptForCommit removes the tag and special characters from the prompt
func CleanPromptForCommit(prompt string) string {
	// List of known tags to strip
	tags := []string{"/bdoc-quick", "/bdoc-engineer", "/bdoc-update", "/bdoc-idea"}
	msg := prompt
	for _, tag := range tags {
		if strings.HasPrefix(msg, tag+" ") {
			msg = strings.TrimPrefix(msg, tag+" ")
			break
		} else if msg == tag {
			msg = ""
			break
		}
	}

	// Keep only alphanumeric characters and spaces
	var b strings.Builder
	for _, r := range msg {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' {
			b.WriteRune(r)
		}
	}

	cleaned := strings.TrimSpace(b.String())
	if len(cleaned) > 255 {
		cleaned = cleaned[:255]
	}
	return cleaned
}

// Duration returns a string representation of the job's duration, rounded to the nearest second
func (j Job) Duration() string {
	if j.Status == "pending" {
		return "scheduled"
	}
	if j.StartedAt == nil {
		return ""
	}
	// If the job is in a terminal state, show the final duration.
	if j.IsTerminal() && j.CompletedAt != nil {
		return j.CompletedAt.Sub(*j.StartedAt).Round(time.Second).String()
	}
	// Defensive: if we are in a terminal state but somehow CompletedAt is missing,
	// don't show a ticking duration.
	if j.IsTerminal() {
		return "interrupted"
	}
	return time.Since(*j.StartedAt).Round(time.Second).String()
}

// ShortCommit returns the first 10 characters of the related commit hash
func (j Job) ShortCommit() string {
	if len(j.RelatedCommit) > 10 {
		return j.RelatedCommit[:10]
	}
	return j.RelatedCommit
}
