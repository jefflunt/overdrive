package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHandleCreateTodo_Errors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-create-todo-errors-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	// Create a Jira project to test disabled creation
	os.MkdirAll("projects/jiraproj", 0755)
	os.WriteFile("projects/jiraproj/project.yml", []byte("name: jiraproj\ntodo_provider: jira\n"), 0644)

	// Create a valid project for json check
	os.MkdirAll("projects/validproj", 0755)
	os.WriteFile("projects/validproj/project.yml", []byte("name: validproj\n"), 0644)
	os.WriteFile("projects/validproj/todos.json", []byte("[]"), 0644)

	tests := []struct {
		name       string
		method     string
		url        string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "Invalid Method",
			method:     http.MethodGet,
			url:        "/projects/p1/todos",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "Missing Project Name",
			method:     http.MethodPost,
			url:        "/todos",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Project Not Found",
			method:     http.MethodPost,
			url:        "/projects/nonexistent/todos",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Jira Project (Creation Disabled)",
			method:     http.MethodPost,
			url:        "/projects/jiraproj/todos",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Invalid JSON",
			method:     http.MethodPost,
			url:        "/projects/validproj/todos",
			body:       "invalid-json",
			wantStatus: http.StatusBadRequest,
		},
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

			HandleCreateTodo(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandleUpdateTodo_Errors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-update-todo-errors-")
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
	os.WriteFile("projects/p1/todos.json", []byte(`[{"id": "t1", "title": "t1", "status": "submitted"}, {"id": "t2", "title": "t2", "status": "draft"}]`), 0644)

	os.MkdirAll("projects/jiraproj", 0755)
	os.WriteFile("projects/jiraproj/project.yml", []byte("name: jiraproj\ntodo_provider: jira\n"), 0644)

	tests := []struct {
		name       string
		method     string
		url        string
		body       interface{}
		wantStatus int
	}{
		{"Invalid Method", http.MethodPost, "/projects/p1/todos/t1", nil, http.StatusMethodNotAllowed},
		{"Missing Project", http.MethodPut, "/todos/t1", nil, http.StatusBadRequest},
		{"Project Not Found", http.MethodPut, "/projects/nonexistent/todos/t1", nil, http.StatusNotFound},
		{"Jira Project", http.MethodPut, "/projects/jiraproj/todos/t1", nil, http.StatusForbidden},
		{"Invalid URL", http.MethodPut, "/projects/p1/todos", nil, http.StatusBadRequest}, // Missing ID
		{"Invalid JSON", http.MethodPut, "/projects/p1/todos/t2", "invalid", http.StatusBadRequest},
		{"Todo Not Found", http.MethodPut, "/projects/p1/todos/nonexistent", map[string]string{"title": "new"}, http.StatusNotFound},
		{"Submitted Todo", http.MethodPut, "/projects/p1/todos/t1", map[string]string{"title": "new"}, http.StatusForbidden},
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

			HandleUpdateTodo(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandleDeleteTodo_Errors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-delete-todo-errors-")
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
	os.WriteFile("projects/p1/todos.json", []byte(`[]`), 0644)

	os.MkdirAll("projects/jiraproj", 0755)
	os.WriteFile("projects/jiraproj/project.yml", []byte("name: jiraproj\ntodo_provider: jira\n"), 0644)

	tests := []struct {
		name       string
		method     string
		url        string
		wantStatus int
	}{
		{"Invalid Method", http.MethodPost, "/projects/p1/todos/t1", http.StatusMethodNotAllowed},
		{"Missing Project", http.MethodDelete, "/todos/t1", http.StatusBadRequest},
		{"Project Not Found", http.MethodDelete, "/projects/nonexistent/todos/t1", http.StatusNotFound},
		{"Jira Project", http.MethodDelete, "/projects/jiraproj/todos/t1", http.StatusForbidden},
		{"Invalid URL", http.MethodDelete, "/projects/p1/todos", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			HandleDeleteTodo(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandleSubmitTodo_Errors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-submit-todo-errors-")
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
	// t1: submitted, t2: draft parent, t3: draft child
	os.WriteFile("projects/p1/todos.json", []byte(`[
        {"id": "t1", "title": "t1", "status": "submitted"},
        {"id": "t2", "title": "t2", "status": "draft", "children": [{"id": "t3", "title": "t3", "status": "draft"}]}
    ]`), 0644)

	tests := []struct {
		name       string
		method     string
		url        string
		wantStatus int
	}{
		{"Invalid Method", http.MethodGet, "/projects/p1/todos/t1/submit", http.StatusMethodNotAllowed},
		{"Missing Project", http.MethodPost, "/todos/t1/submit", http.StatusBadRequest},
		{"Project Not Found", http.MethodPost, "/projects/nonexistent/todos/t1/submit", http.StatusNotFound},
		{"Invalid URL", http.MethodPost, "/projects/p1/todos", http.StatusBadRequest}, // Missing submit or id
		{"Todo Not Found", http.MethodPost, "/projects/p1/todos/nonexistent/submit", http.StatusNotFound},
		{"Submit Parent", http.MethodPost, "/projects/p1/todos/t2/submit", http.StatusBadRequest},
		{"Submit Non-Draft", http.MethodPost, "/projects/p1/todos/t1/submit", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			HandleSubmitTodo(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
