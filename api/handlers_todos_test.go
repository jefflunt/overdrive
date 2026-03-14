package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleListTodos(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-todos-")
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
	os.MkdirAll(filepath.Join("projects", projectName), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "todos.json"), []byte(`[{"id": "t1", "title": "todo 1"}]`), 0644)

	req := httptest.NewRequest("GET", "/projects/testproject/todos", nil)
	w := httptest.NewRecorder()

	HandleListTodos(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d: %s", resp.StatusCode, bodyStr)
	}

	if !strings.Contains(bodyStr, "todo 1") {
		t.Errorf("Expected body to contain todo 1, got: %s", bodyStr)
	}
}

func TestHandleCreateTodo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-create-todo-")
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
	os.MkdirAll(filepath.Join("projects", projectName), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)

	todo := Todo{Title: "New Todo", Description: "New Description"}
	body, _ := json.Marshal(todo)
	req := httptest.NewRequest("POST", "/projects/testproject/todos", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	HandleCreateTodo(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var created Todo
	json.NewDecoder(resp.Body).Decode(&created)
	if created.Title != "New Todo" {
		t.Errorf("Expected title New Todo, got %s", created.Title)
	}
}

func TestHandleUpdateTodo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-update-todo-")
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
	os.MkdirAll(filepath.Join("projects", projectName), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "todos.json"), []byte(`[{"id": "t1", "title": "todo 1", "status": "draft"}]`), 0644)

	update := Todo{ID: "t1", Title: "updated title"}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest("PUT", "/projects/testproject/todos/t1", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	HandleUpdateTodo(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Verify update
	todosData, _ := os.ReadFile(filepath.Join("projects", projectName, "todos.json"))
	if !strings.Contains(string(todosData), "updated title") {
		t.Error("Expected todos.json to be updated")
	}
}

func TestHandleDeleteTodo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-delete-todo-")
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
	os.MkdirAll(filepath.Join("projects", projectName), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "todos.json"), []byte(`[{"id": "t1", "title": "todo 1"}]`), 0644)

	req := httptest.NewRequest("DELETE", "/projects/testproject/todos/t1", nil)
	w := httptest.NewRecorder()

	HandleDeleteTodo(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Verify deletion
	todosData, _ := os.ReadFile(filepath.Join("projects", projectName, "todos.json"))
	if strings.Contains(string(todosData), "t1") {
		t.Error("Expected todo to be deleted")
	}
}

func TestHandleSubmitTodo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-submit-todo-")
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
	os.WriteFile(filepath.Join("projects", projectName, "todos.json"), []byte(`[{"id": "t1", "title": "todo 1", "status": "draft"}]`), 0644)

	req := httptest.NewRequest("POST", "/projects/testproject/todos/t1/submit", nil)
	w := httptest.NewRecorder()

	HandleSubmitTodo(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Verify job creation
	files, _ := filepath.Glob(filepath.Join("projects", projectName, "jobs", "pending", "*.yml"))
	if len(files) != 1 {
		t.Errorf("Expected 1 job file, got %d", len(files))
	}

	// Verify todo status update
	todosData, _ := os.ReadFile(filepath.Join("projects", projectName, "todos.json"))
	if !strings.Contains(string(todosData), "submitted") {
		t.Error("Expected todo status to be submitted")
	}
}
