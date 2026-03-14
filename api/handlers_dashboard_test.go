package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestHandleGetDashboardTodos(t *testing.T) {
	// Setup test environment
	tmpDir, err := os.MkdirTemp("", "dashboard_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create a project
	projectName := "testproj"
	project := &Project{
		Name: projectName,
	}
	if err := SaveProject(project); err != nil {
		t.Fatal(err)
	}

	// Create a working job
	jobID := "job1"
	job := Job{
		ID:        jobID,
		Project:   projectName,
		Status:    "working",
		CreatedAt: time.Now(),
		Request: JobRequest{
			Prompt: "Test job",
		},
	}
	jobData, _ := yaml.Marshal(job)
	jobPath := filepath.Join(project.JobsPath("working"), jobID+".yml")
	os.WriteFile(jobPath, jobData, 0644)

	// Create a pending job
	jobID2 := "job2"
	job2 := Job{
		ID:        jobID2,
		Project:   projectName,
		Status:    "pending",
		CreatedAt: time.Now().Add(time.Minute),
		Request: JobRequest{
			Prompt: "Test job 2",
		},
	}
	jobData2, _ := yaml.Marshal(job2)
	jobPath2 := filepath.Join(project.JobsPath("pending"), jobID2+".yml")
	os.WriteFile(jobPath2, jobData2, 0644)

	// Create a todo
	todos := []Todo{
		{
			ID:      "todo1",
			Title:   "Test Todo",
			Starred: true,
		},
	}
	SaveTodos(projectName, todos)

	// Request dashboard data
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/todos", nil)
	w := httptest.NewRecorder()
	HandleGetDashboardTodos(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", w.Code)
	}

	var data DashboardData
	if err := json.NewDecoder(w.Body).Decode(&data); err != nil {
		t.Fatal(err)
	}

	// Verify running jobs
	if len(data.RunningJobs) != 2 {
		t.Errorf("expected 2 running jobs, got %v", len(data.RunningJobs))
	}

	// Verify jobs are sorted by CreatedAt descending
	if data.RunningJobs[0].ID != jobID2 {
		t.Errorf("expected first job to be %v, got %v", jobID2, data.RunningJobs[0].ID)
	}

	// Verify todos
	if len(data.Projects[projectName]) != 1 {
		t.Errorf("expected 1 todo for project, got %v", len(data.Projects[projectName]))
	}
}

func TestHandleGetDashboardRecentJobsLimit(t *testing.T) {
	// Setup test environment
	tmpDir, err := os.MkdirTemp("", "dashboard_limit_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create a project
	projectName := "testproj"
	project := &Project{
		Name: projectName,
	}
	if err := SaveProject(project); err != nil {
		t.Fatal(err)
	}

	// Create 10 finished jobs
	for i := 0; i < 10; i++ {
		jobID := "job" + string(rune('a'+i))
		job := Job{
			ID:        jobID,
			Project:   projectName,
			Status:    "done",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
		}
		jobData, _ := yaml.Marshal(job)
		jobPath := filepath.Join(project.JobsPath("done"), jobID+".yml")
		os.WriteFile(jobPath, jobData, 0644)
	}

	// Request dashboard data
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/todos", nil)
	w := httptest.NewRecorder()
	HandleGetDashboardTodos(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", w.Code)
	}

	var data DashboardData
	if err := json.NewDecoder(w.Body).Decode(&data); err != nil {
		t.Fatal(err)
	}

	// Verify recent jobs limit
	if len(data.RecentJobs) != 7 {
		t.Errorf("expected 7 recent jobs, got %v", len(data.RecentJobs))
	}
}
