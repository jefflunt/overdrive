package api

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestEnqueueJob_ModelDefaults(t *testing.T) {
	// Setup test project
	projectName := "TestWorkerModelDefaults"
	project := Project{
		Name:    projectName,
		RepoURL: "https://github.com/example/repo",
	}

	// Clean up
	_ = DeleteProject(projectName)
	defer DeleteProject(projectName)

	if err := SaveProject(&project); err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}

	// Test 1: No project model, no request model -> default flash
	req1 := JobRequest{
		Prompt: "Test 1",
	}
	jobID1, err := EnqueueJob(project, req1)
	if err != nil {
		t.Fatalf("Failed to enqueue job 1: %v", err)
	}

	job1 := loadJob(t, project, jobID1)
	if job1.Request.Model != "google/gemini-3-flash-preview" {
		t.Errorf("Job 1: Expected model 'google/gemini-3-flash-preview', got '%s'", job1.Request.Model)
	}

	// Test 2: Project model set, no request model -> project model
	project.BuildModel = "google/gemini-3-pro-preview"
	if err := SaveProject(&project); err != nil {
		t.Fatalf("Failed to save project with model: %v", err)
	}

	req2 := JobRequest{
		Prompt: "Test 2",
	}
	jobID2, err := EnqueueJob(project, req2)
	if err != nil {
		t.Fatalf("Failed to enqueue job 2: %v", err)
	}

	job2 := loadJob(t, project, jobID2)
	if job2.Request.Model != "google/gemini-3-pro-preview" {
		t.Errorf("Job 2: Expected model 'google/gemini-3-pro-preview', got '%s'", job2.Request.Model)
	}

	// Test 3: Project model set, request model set -> request model
	req3 := JobRequest{
		Prompt: "Test 3",
		Model:  "google/gemini-2.0-flash-thinking-exp-1219", // Some other model
	}
	jobID3, err := EnqueueJob(project, req3)
	if err != nil {
		t.Fatalf("Failed to enqueue job 3: %v", err)
	}

	job3 := loadJob(t, project, jobID3)
	if job3.Request.Model != "google/gemini-2.0-flash-thinking-exp-1219" {
		t.Errorf("Job 3: Expected model 'google/gemini-2.0-flash-thinking-exp-1219', got '%s'", job3.Request.Model)
	}
}

func loadJob(t *testing.T, project Project, jobID string) *Job {
	path := filepath.Join(project.JobsPath("pending"), jobID+".yml")
	// Retry a few times just in case file system is slow (unlikely for pending)
	var data []byte
	var err error
	for i := 0; i < 3; i++ {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("Failed to read job file %s: %v", path, err)
	}

	var job Job
	if err := yaml.Unmarshal(data, &job); err != nil {
		t.Fatalf("Failed to unmarshal job %s: %v", path, err)
	}
	return &job
}
