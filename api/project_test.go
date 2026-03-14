package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectPaths(t *testing.T) {
	p := Project{Name: "myproj"}
	if p.Path() != filepath.Join("projects", "myproj") {
		t.Errorf("Path() = %v", p.Path())
	}
	if p.JobsPath("pending") != filepath.Join("projects", "myproj", "jobs", "pending") {
		t.Errorf("JobsPath() = %v", p.JobsPath("pending"))
	}
	if p.LogsPath("job123") != filepath.Join("projects", "myproj", "logs", "job123") {
		t.Errorf("LogsPath() = %v", p.LogsPath("job123"))
	}
	if p.ChatsPath() != filepath.Join("projects", "myproj", "chats") {
		t.Errorf("ChatsPath() = %v", p.ChatsPath())
	}
}

func TestProjectManagement(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "project_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	// Ensure 'projects' dir exists
	if err := os.Mkdir("projects", 0755); err != nil {
		t.Fatal(err)
	}

	// Test ListProjects (empty)
	projects, err := ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}

	// Test SaveProject
	p := &Project{
		Name:    "test-proj",
		RepoURL: "https://github.com/test/repo",
	}
	err = SaveProject(p)
	if err != nil {
		t.Fatalf("SaveProject failed: %v", err)
	}

	// Verify dirs created
	if _, err := os.Stat("projects/test-proj/jobs/pending"); err != nil {
		t.Errorf("pending job dir not created: %v", err)
	}

	// Test GetProject
	gp, err := GetProject("test-proj")
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if gp.Name != "test-proj" {
		t.Errorf("expected test-proj, got %s", gp.Name)
	}
	if gp.TodoProvider != "native" {
		t.Errorf("expected native TodoProvider, got %s", gp.TodoProvider)
	}

	// Test migration of LegacyModel
	if err := os.MkdirAll("projects/legacy", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("projects/legacy/project.yml", []byte("model: my-model\n"), 0644); err != nil {
		t.Fatal(err)
	}
	lp, err := GetProject("legacy")
	if err != nil {
		t.Fatalf("GetProject legacy failed: %v", err)
	}
	if lp.BuildModel != "my-model" || lp.ChatModel != "my-model" {
		t.Errorf("migration failed: BuildModel=%s, ChatModel=%s", lp.BuildModel, lp.ChatModel)
	}
	if lp.Build.LLM.Model != "my-model" || lp.Chat.LLM.Model != "my-model" {
		t.Errorf("new config migration failed: Build.LLM.Model=%s, Chat.LLM.Model=%s", lp.Build.LLM.Model, lp.Chat.LLM.Model)
	}
	if lp.Build.Harness.Provider != "opencode" || lp.Chat.Harness.Provider != "opencode" {
		t.Errorf("default harness migration failed: Build.Harness.Provider=%s", lp.Build.Harness.Provider)
	}

	// Test Pause/Resume
	err = PauseProject("test-proj")
	if err != nil {
		t.Fatal(err)
	}
	gp, _ = GetProject("test-proj")
	if !gp.Paused {
		t.Error("expected paused=true")
	}

	err = ResumeProject("test-proj")
	if err != nil {
		t.Fatal(err)
	}
	gp, _ = GetProject("test-proj")
	if gp.Paused {
		t.Error("expected paused=false")
	}

	// Test ListProjects (not empty)
	projects, err = ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	// legacy and test-proj
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}

	// Test SSH Key/Config handling
	p.SSHKey = "SECRET KEY"
	p.SSHConfig = "CONFIG CONTENT"
	err = SaveProject(p)
	if err != nil {
		t.Fatal(err)
	}

	gp, _ = GetProject("test-proj")
	// GetProject should return absolute paths if files exist
	if !filepath.IsAbs(gp.SSHKey) || !filepath.IsAbs(gp.SSHConfig) {
		t.Errorf("expected absolute paths for SSHKey/SSHConfig, got %s / %s", gp.SSHKey, gp.SSHConfig)
	}

	// Test DeleteProject
	err = DeleteProject("test-proj")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat("projects/test-proj"); !os.IsNotExist(err) {
		t.Error("project dir still exists after delete")
	}
}
