package api

import (
	"strings"
	"testing"
	"time"
)

func TestJob_Duration(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		startedAt   *time.Time
		completedAt *time.Time
		status      string
		expected    string
	}{
		{
			name:        "nil started at",
			startedAt:   nil,
			completedAt: &now,
			expected:    "",
		},
		{
			name:        "pending job",
			startedAt:   nil,
			completedAt: nil,
			status:      "pending",
			expected:    "scheduled",
		},
		{
			name:        "active job (nil completed at)",
			startedAt:   &now,
			completedAt: nil,
			status:      "working",
			expected:    "0s",
		},
		{
			name:        "both nil",
			startedAt:   nil,
			completedAt: nil,
			expected:    "",
		},
		{
			name:        "exactly 5 seconds",
			startedAt:   &now,
			completedAt: pointer(now.Add(5 * time.Second)),
			status:      "done",
			expected:    "5s",
		},
		{
			name:        "5 seconds and 500ms (round down)",
			startedAt:   &now,
			completedAt: pointer(now.Add(5*time.Second + 499*time.Millisecond)),
			status:      "done",
			expected:    "5s",
		},
		{
			name:        "5 seconds and 501ms (round up)",
			startedAt:   &now,
			completedAt: pointer(now.Add(5*time.Second + 501*time.Millisecond)),
			status:      "done",
			expected:    "6s",
		},
		{
			name:        "1 minute 5 seconds",
			startedAt:   &now,
			completedAt: pointer(now.Add(1*time.Minute + 5*time.Second)),
			status:      "done",
			expected:    "1m5s",
		},
		{
			name:        "1 hour 2 minutes 3 seconds",
			startedAt:   &now,
			completedAt: pointer(now.Add(1*time.Hour + 2*time.Minute + 3*time.Second)),
			status:      "done",
			expected:    "1h2m3s",
		},
		{
			name:        "crashed job with nil completed at",
			startedAt:   &now,
			completedAt: nil,
			status:      "crash",
			expected:    "interrupted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := Job{
				StartedAt:   tt.startedAt,
				CompletedAt: tt.completedAt,
				Status:      tt.status,
			}
			if got := j.Duration(); got != tt.expected {
				t.Errorf("Job.Duration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateJobID(t *testing.T) {
	id1 := GenerateJobID()
	time.Sleep(1 * time.Millisecond)
	id2 := GenerateJobID()

	if len(id1) < 10 {
		t.Errorf("ID too short: %s", id1)
	}

	if id1 >= id2 {
		t.Errorf("IDs not sortable: %s should be less than %s", id1, id2)
	}

	// Check alphanumeric
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for _, char := range id1 {
		found := false
		for _, c := range chars {
			if char == c {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ID contains non-alphanumeric character: %c in %s", char, id1)
		}
	}
}

func TestJob_IsTerminal(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"pending", false},
		{"working", false},
		{"done", true},
		{"crash", true},
		{"no-op", true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			j := Job{Status: tt.status}
			if got := j.IsTerminal(); got != tt.expected {
				t.Errorf("Job.IsTerminal() for status %v = %v, want %v", tt.status, got, tt.expected)
			}
		})
	}
}

func TestCleanPromptForCommit(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		expected string
	}{
		{
			name:     "no tag",
			prompt:   "fix the bug",
			expected: "fix the bug",
		},
		{
			name:     "bdoc-quick tag",
			prompt:   "/bdoc-quick fix the bug",
			expected: "fix the bug",
		},
		{
			name:     "special characters",
			prompt:   "/bdoc-quick fix bug #123! @user",
			expected: "fix bug 123 user",
		},
		{
			name:     "too long",
			prompt:   "/bdoc-quick " + strings.Repeat("a", 300),
			expected: strings.Repeat("a", 255),
		},
		{
			name:     "only tag",
			prompt:   "/bdoc-quick",
			expected: "",
		},
		{
			name:     "multiple spaces",
			prompt:   "/bdoc-quick   fix   bug   ",
			expected: "fix   bug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanPromptForCommit(tt.prompt); got != tt.expected {
				t.Errorf("CleanPromptForCommit() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func pointer(t time.Time) *time.Time {
	return &t
}
