package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHandleSaveProject_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/projects/save", nil)
	w := httptest.NewRecorder()

	HandleSaveProject(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleSaveProject_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/projects/save", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	HandleSaveProject(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleSaveProject_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		project Project
	}{
		{"Missing RepoURL", Project{Name: "p1", SSHKey: "key", PrimaryBranch: "main"}},
		{"Missing SSHKey", Project{Name: "p1", RepoURL: "repo", PrimaryBranch: "main"}},
		{"Missing PrimaryBranch", Project{Name: "p1", RepoURL: "repo", SSHKey: "key"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.project)
			req := httptest.NewRequest(http.MethodPost, "/projects/save", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			HandleSaveProject(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestHandleGetProjectConfig_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-get-project-config-notfound-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	req := httptest.NewRequest(http.MethodGet, "/projects/nonexistent/config", nil)
	w := httptest.NewRecorder()

	HandleGetProjectConfig(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleDeleteProject_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/projects/p1", nil)
	w := httptest.NewRecorder()

	HandleDeleteProject(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleGetProjectConfig_MissingName(t *testing.T) {
	// If path doesn't contain projects/name, getProjectName returns ""
	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	w := httptest.NewRecorder()

	HandleGetProjectConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
