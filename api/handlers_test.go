package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestGetProjectName(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "query param",
			url:      "/some/path?project=myproj",
			expected: "myproj",
		},
		{
			name:     "path param",
			url:      "/projects/myproj/jobs",
			expected: "myproj",
		},
		{
			name:     "no project",
			url:      "/other/path",
			expected: "",
		},
		{
			name:     "projects at end",
			url:      "/some/projects",
			expected: "",
		},
		{
			name:     "query param overrides path",
			url:      "/projects/pathproj/jobs?project=queryproj",
			expected: "queryproj",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			got := getProjectName(req)
			if got != tt.expected {
				t.Errorf("getProjectName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseTemplate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	// Test parsing layout only
	tmpl, err := parseTemplate("")
	if err != nil {
		t.Fatalf("parseTemplate(\"\") failed: %v", err)
	}
	if tmpl == nil {
		t.Fatal("parseTemplate(\"\") returned nil template")
	}

	// Test custom functions
	t.Run("template functions", func(t *testing.T) {
		tests := []struct {
			name     string
			template string
			data     interface{}
			want     string
		}{
			{
				name:     "add",
				template: `{{add 1 2}}`,
				want:     "3",
			},
			{
				name:     "dict",
				template: `{{index (dict "a" 1 "b" "two") "b"}}`,
				want:     "two",
			},
			{
				name:     "dict invalid args",
				template: `{{dict "a"}}`,
				want:     "ERROR",
			},
			{
				name:     "dict non-string key",
				template: `{{dict 1 "a"}}`,
				want:     "ERROR",
			},
			{
				name:     "formatDate",
				template: `{{formatDate .}}`,
				data:     time.Date(2023, 10, 27, 10, 11, 12, 0, time.UTC),
				want:     "2023-10-27 10:11:12",
			},
			{
				name:     "json",
				template: `{{json .}}`,
				data:     map[string]string{"foo": "bar"},
				want:     `{&#34;foo&#34;:&#34;bar&#34;}`,
			},
			{
				name:     "getBuildNumber",
				template: `{{getBuildNumber}}`,
				want:     "b0", // In temp dir without git, wc -l returns 0
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				nt, err := tmpl.Clone()
				if err != nil {
					t.Fatal(err)
				}
				nt, err = nt.New("test").Parse(tt.template)
				if err != nil {
					t.Fatalf("failed to parse test template: %v", err)
				}
				var buf bytes.Buffer
				err = nt.Execute(&buf, tt.data)
				if tt.want == "ERROR" {
					if err == nil {
						t.Errorf("expected error but got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("execution failed: %v", err)
				}
				if buf.String() != tt.want {
					t.Errorf("got %q, want %q", buf.String(), tt.want)
				}
			})
		}
	})
}
