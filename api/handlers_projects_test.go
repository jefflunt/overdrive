package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleListProjects(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-list-projects-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	os.MkdirAll("projects/project1", 0755)
	os.MkdirAll("projects/project2", 0755)
	os.WriteFile("projects/project1/project.yml", []byte("name: project1\n"), 0644)
	os.WriteFile("projects/project2/project.yml", []byte("name: project2\n"), 0644)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	HandleListProjects(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "project1") || !strings.Contains(bodyStr, "project2") {
		t.Errorf("Expected body to contain projects")
	}
}

func TestHandleGetProjectConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-get-project-")
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
	projectYaml := []byte(fmt.Sprintf("name: %s\nrepo_url: https://github.com/example/repo.git\n", projectName))
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), projectYaml, 0644)

	req := httptest.NewRequest("GET", fmt.Sprintf("/projects/%s/config", projectName), nil)
	w := httptest.NewRecorder()

	HandleGetProjectConfig(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		t.Fatal(err)
	}

	if project.Name != projectName {
		t.Errorf("Expected project name %s, got %s", projectName, project.Name)
	}
}

func TestHandleSaveProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-save-project-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	project := Project{
		Name:          "newproject",
		RepoURL:       "https://github.com/example/new.git",
		SSHKey:        "testkey",
		PrimaryBranch: "main",
	}
	body, _ := json.Marshal(project)
	req := httptest.NewRequest("POST", "/projects/save", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	HandleSaveProject(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Verify project saved
	_, err = GetProject("newproject")
	if err != nil {
		t.Errorf("Project not found after save: %v", err)
	}
}

func TestHandleDeleteProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-delete-project-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	projectName := "project-to-delete"
	os.MkdirAll(filepath.Join("projects", projectName), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)

	req := httptest.NewRequest("DELETE", "/projects/"+projectName, nil)
	w := httptest.NewRecorder()

	HandleDeleteProject(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Verify project deleted
	if _, err := os.Stat(filepath.Join("projects", projectName)); !os.IsNotExist(err) {
		t.Error("Expected project directory to be deleted")
	}
}
