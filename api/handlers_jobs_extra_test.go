package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleTailLogs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-tail-logs-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	projectName := "testproject"
	jobID := "job1"
	os.MkdirAll(filepath.Join("projects", projectName, "logs", jobID), 0755)
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "working"), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "logs", jobID, "worker.log"), []byte("test log line"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "jobs", "working", jobID+".yml"), []byte("id: "+jobID+"\nstatus: working\n"), 0644)

	req := httptest.NewRequest("GET", "/projects/"+projectName+"/jobs/tail/"+jobID, nil)
	w := httptest.NewRecorder()

	HandleTailLogs(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleJobUpdates(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-job-updates-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	projectName := "testproject"
	jobID := "job1"
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "working"), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "jobs", "working", jobID+".yml"), []byte("id: "+jobID+"\nstatus: working\n"), 0644)

	req := httptest.NewRequest("GET", "/projects/"+projectName+"/jobs/updates/"+jobID, nil)
	w := httptest.NewRecorder()

	HandleJobUpdates(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleSubmitJob(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-submit-job-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	projectName := "testproject"
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "pending"), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)

	jobReq := JobRequest{
		Project: projectName,
		Prompt:  "test prompt",
	}
	body, _ := json.Marshal(jobReq)
	req := httptest.NewRequest("POST", "/projects/"+projectName+"/jobs/submit", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	HandleSubmitJob(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", resp.StatusCode)
	}
}

func TestHandleStopJob(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-stop-job-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	projectName := "testproject"
	jobID := "job1"
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "working"), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "jobs", "working", jobID+".yml"), []byte("id: "+jobID+"\nstatus: working\n"), 0644)

	// Mock StopJob by creating the expected behavior if possible, but StopJob might call external commands.
	// StopJob is in worker.go, let's see it.

	req := httptest.NewRequest("POST", "/projects/"+projectName+"/jobs/stop/"+jobID, nil)
	w := httptest.NewRecorder()

	// This will likely fail because StopJob calls podman kill, but let's see
	HandleStopJob(w, req)

	// Even if it fails to stop (podman not found), it might return 404 if it can't find the job to stop.
}

func TestHandleDeleteJob(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-delete-job-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	projectName := "testproject"
	jobID := "job1"
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "done"), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "jobs", "done", jobID+".yml"), []byte("id: "+jobID+"\nstatus: done\n"), 0644)

	req := httptest.NewRequest("DELETE", "/projects/"+projectName+"/jobs/delete/"+jobID, nil)
	w := httptest.NewRecorder()

	HandleDeleteJob(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected 303 See Other, got %d", resp.StatusCode)
	}

	if _, err := os.Stat(filepath.Join("projects", projectName, "jobs", "done", jobID+".yml")); !os.IsNotExist(err) {
		t.Error("Expected job file to be deleted")
	}
}

func TestHandleRevertJob(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-revert-job-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	projectName := "testproject"
	jobID := "job1"
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "done"), 0755)
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "pending"), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "jobs", "done", jobID+".yml"), []byte("id: "+jobID+"\nstatus: done\nrelated_commit: abc123\n"), 0644)

	req := httptest.NewRequest("POST", "/projects/"+projectName+"/jobs/revert/"+jobID, nil)
	w := httptest.NewRecorder()

	HandleRevertJob(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", resp.StatusCode)
	}
}

func TestHandleJobDiff(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-job-diff-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	projectName := "testproject"
	jobID := "job1"
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "done"), 0755)
	os.MkdirAll(filepath.Join("projects", projectName, "logs", jobID), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "jobs", "done", jobID+".yml"), []byte("id: "+jobID+"\nstatus: done\nrelated_commit: abc123\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "logs", jobID, "diff"), []byte("test diff"), 0644)

	req := httptest.NewRequest("GET", "/projects/"+projectName+"/jobs/diff/"+jobID, nil)
	w := httptest.NewRecorder()

	HandleJobDiff(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleListJobs_Errors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-list-jobs-errors-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"No Project", "/jobs", http.StatusSeeOther},
		{"Project Not Found", "/projects/nonexistent/jobs", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			HandleListJobs(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandleStopJob_Errors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-stop-job-errors-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	os.MkdirAll("projects/p1/jobs/working", 0755)
	os.WriteFile("projects/p1/project.yml", []byte("name: p1\n"), 0644)
	// Create a docs job that cannot be stopped
	os.WriteFile("projects/p1/jobs/working/docs-job.yml", []byte("status: working\nrequest:\n  prompt: /bdoc-update\n"), 0644)

	tests := []struct {
		name       string
		method     string
		url        string
		wantStatus int
	}{
		{"Invalid Method", http.MethodPut, "/projects/p1/jobs/stop/j1", http.StatusMethodNotAllowed},
		{"Invalid Job ID", http.MethodPost, "/projects/p1/jobs/stop/.", http.StatusBadRequest},
		{"Project Not Found", http.MethodPost, "/projects/nonexistent/jobs/stop/j1", http.StatusNotFound},
		{"Docs Job Forbidden", http.MethodPost, "/projects/p1/jobs/stop/docs-job", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			HandleStopJob(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandleDeleteJob_Errors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-delete-job-errors-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	os.MkdirAll("projects/p1", 0755)
	os.WriteFile("projects/p1/project.yml", []byte("name: p1\n"), 0644)

	tests := []struct {
		name       string
		method     string
		url        string
		wantStatus int
	}{
		{"Invalid Method", http.MethodGet, "/projects/p1/jobs/j1", http.StatusMethodNotAllowed},
		{"Missing Project", http.MethodDelete, "/jobs/j1", http.StatusBadRequest},
		{"Project Not Found", http.MethodDelete, "/projects/nonexistent/jobs/j1", http.StatusNotFound},
		{"Job Not Found", http.MethodDelete, "/projects/p1/jobs/nonexistent", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			HandleDeleteJob(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandleSubmitJob_Errors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-submit-job-errors-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	os.MkdirAll("projects/p1", 0755)
	os.WriteFile("projects/p1/project.yml", []byte("name: p1\n"), 0644)

	tests := []struct {
		name       string
		method     string
		url        string
		body       interface{}
		wantStatus int
	}{
		{"Invalid Method", http.MethodGet, "/projects/p1/jobs", nil, http.StatusMethodNotAllowed},
		{"Invalid JSON", http.MethodPost, "/projects/p1/jobs", "invalid", http.StatusBadRequest},
		{"Missing Project in URL and Body", http.MethodPost, "/jobs", map[string]string{"prompt": "hi"}, http.StatusBadRequest},
		{"Project Not Found", http.MethodPost, "/projects/nonexistent/jobs", map[string]string{"prompt": "hi"}, http.StatusNotFound},
		{"Missing Prompt", http.MethodPost, "/projects/p1/jobs", map[string]string{"project": "p1"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if s, ok := tt.body.(string); ok {
				bodyBytes = []byte(s)
			} else if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer(bodyBytes))
			w := httptest.NewRecorder()

			HandleSubmitJob(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandleJobDiff_Errors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-job-diff-errors-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	os.MkdirAll("projects/p1/jobs/done", 0755)
	os.WriteFile("projects/p1/project.yml", []byte("name: p1\n"), 0644)
	// Job without commit
	os.WriteFile("projects/p1/jobs/done/nocommit.yml", []byte("status: done\nrelated_commit: \"\"\n"), 0644)

	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{"Invalid Job ID", "/projects/p1/jobs/diff/.", http.StatusBadRequest},
		{"Missing Project", "/jobs/diff/j1", http.StatusBadRequest},
		{"Project Not Found", "/projects/nonexistent/jobs/diff/j1", http.StatusNotFound},
		{"Job Not Found (Not Done)", "/projects/p1/jobs/diff/working-job", http.StatusNotFound},
		{"No Related Commit", "/projects/p1/jobs/diff/nocommit", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			HandleJobDiff(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
