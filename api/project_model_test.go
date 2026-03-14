package api

import (
	"testing"
)

func TestProjectModelPersistence(t *testing.T) {
	// Setup
	projectName := "TestModelPersistence"
	project := Project{
		Name:       projectName,
		RepoURL:    "https://github.com/example/repo",
		BuildModel: "google/gemini-3-pro-preview",
	}

	// Clean up any existing test project
	_ = DeleteProject(projectName)
	defer DeleteProject(projectName)

	// Save Project
	err := SaveProject(&project)
	if err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}

	// Load Project
	loadedProject, err := GetProject(projectName)
	if err != nil {
		t.Fatalf("Failed to load project: %v", err)
	}

	// Verify Model
	if loadedProject.BuildModel != "google/gemini-3-pro-preview" {
		t.Errorf("Expected model 'google/gemini-3-pro-preview', got '%s'", loadedProject.BuildModel)
	}

	// Test Cmds
	project.Cmds = []ProjectCmd{
		{Label: "Build", Cmd: "npm run build"},
		{Label: "Test", Cmd: "npm test"},
	}
	err = SaveProject(&project)
	if err != nil {
		t.Fatalf("Failed to save project with cmds: %v", err)
	}

	loadedProject, err = GetProject(projectName)
	if err != nil {
		t.Fatalf("Failed to load project after saving cmds: %v", err)
	}

	if len(loadedProject.Cmds) != 2 {
		t.Errorf("Expected 2 cmds, got %d", len(loadedProject.Cmds))
	} else {
		if loadedProject.Cmds[0].Label != "Build" || loadedProject.Cmds[0].Cmd != "npm run build" {
			t.Errorf("Cmd 0 mismatch: %+v", loadedProject.Cmds[0])
		}
		if loadedProject.Cmds[1].Label != "Test" || loadedProject.Cmds[1].Cmd != "npm test" {
			t.Errorf("Cmd 1 mismatch: %+v", loadedProject.Cmds[1])
		}
	}

	// Test with empty model (should happen for legacy projects)
	project.BuildModel = ""
	err = SaveProject(&project)
	if err != nil {
		t.Fatalf("Failed to save project with empty model: %v", err)
	}

	loadedProject, err = GetProject(projectName)
	if err != nil {
		t.Fatalf("Failed to load project: %v", err)
	}

	if loadedProject.BuildModel != "" {
		t.Errorf("Expected empty model, got '%s'", loadedProject.BuildModel)
	}
}
