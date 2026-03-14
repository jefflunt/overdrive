package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTodoLogic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-todo-logic-")
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

	todos := []Todo{
		{ID: "p1", Title: "Parent 1", Status: "draft"},
	}

	// Add child
	child := Todo{ID: "c1", ParentID: "p1", Title: "Child 1", Status: "draft"}
	updated, err := AddTodo(todos, child)
	if err != nil {
		t.Fatalf("AddTodo failed: %v", err)
	}
	if len(updated[0].Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(updated[0].Children))
	}

	// Find
	found := FindTodo(updated, "c1")
	if found == nil || found.Title != "Child 1" {
		t.Errorf("FindTodo failed to find child")
	}

	// Update status
	updated, ok := UpdateTodoStatus(updated, "c1", "submitted", "job1")
	if !ok {
		t.Fatal("UpdateTodoStatus failed")
	}
	found = FindTodo(updated, "c1")
	if found.Status != "submitted" || found.JobID != "job1" {
		t.Errorf("Expected status submitted and job1, got %s and %s", found.Status, found.JobID)
	}

	// Find by JobID
	found = FindTodoByJobID(updated, "job1")
	if found == nil || found.ID != "c1" {
		t.Errorf("FindTodoByJobID failed")
	}

	// Update in tree
	updateReq := Todo{ID: "c1", Title: "Updated Child"}
	updated, ok = UpdateTodoInTree(updated, updateReq)
	if !ok {
		t.Fatal("UpdateTodoInTree failed")
	}
	found = FindTodo(updated, "c1")
	if found.Title != "Updated Child" {
		t.Errorf("Expected Updated Child, got %s", found.Title)
	}

	// Delete from tree
	updated = DeleteTodoFromTree(updated, "c1")
	if len(updated[0].Children) != 0 {
		t.Error("DeleteTodoFromTree failed")
	}

	// Save and Load
	if err := SaveTodos(projectName, updated); err != nil {
		t.Fatalf("SaveTodos failed: %v", err)
	}
	loaded, err := LoadTodos(projectName)
	if err != nil {
		t.Fatalf("LoadTodos failed: %v", err)
	}
	if len(loaded) != 1 || loaded[0].ID != "p1" {
		t.Error("LoadTodos failed")
	}
}

func TestUpdateTodoStatusFromJob(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-todo-job-")
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

	todos := []Todo{
		{ID: "t1", JobID: "job1", Status: "submitted"},
	}
	SaveTodos(projectName, todos)

	job := &Job{
		ID:      "job1",
		Status:  "done",
		Project: projectName,
	}

	if err := UpdateTodoStatusFromJob(projectName, job); err != nil {
		t.Fatalf("UpdateTodoStatusFromJob failed: %v", err)
	}

	loaded, _ := LoadTodos(projectName)
	if loaded[0].Status != "completed" {
		t.Errorf("Expected status completed, got %s", loaded[0].Status)
	}
}
