package api

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCleanupZombieJobs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-worker-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	projectName := "testproject"
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "working"), 0755)
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "crash"), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)

	jobFile := filepath.Join("projects", projectName, "jobs", "working", "job1.yml")
	os.WriteFile(jobFile, []byte("id: job1\nstatus: working\n"), 0644)

	CleanupZombieJobs()

	if _, err := os.Stat(jobFile); !os.IsNotExist(err) {
		t.Error("Expected job file to be moved from working")
	}

	crashFile := filepath.Join("projects", projectName, "jobs", "crash", "job1.yml")
	if _, err := os.Stat(crashFile); os.IsNotExist(err) {
		t.Error("Expected job file to be moved to crash")
	}
}

func TestJobCount(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-jobcount-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	projectName := "testproject"
	os.MkdirAll(filepath.Join("projects", projectName), 0755)
	project := Project{Name: projectName}

	err = writeJobCount(project, 5)
	if err != nil {
		t.Fatal(err)
	}

	count, err := readJobCount(project)
	if err != nil {
		t.Fatal(err)
	}

	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}
}

func TestSchedule(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-schedule-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	projectName := "testproject"
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "pending"), 0755)
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "working"), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "jobs", "pending", "job1.yml"), []byte("id: job1\nstatus: pending\n"), 0644)

	// To avoid calling processJob which starts a real script, we can't easily call schedule()
	// unless we mock ./scripts/work.

	os.MkdirAll("scripts", 0755)
	os.WriteFile("scripts/work", []byte("#!/bin/sh\nexit 0\n"), 0755)

	schedule()

	// Verify job moved to working
	files, _ := filepath.Glob(filepath.Join("projects", projectName, "jobs", "working", "*.yml"))
	if len(files) != 1 {
		t.Errorf("Expected 1 working job, got %d", len(files))
	}
}

func TestStopJob(t *testing.T) {
	jobID := "stopme"
	cancelled := false
	cancel := func() {
		cancelled = true
	}
	activeContexts.Store(jobID, jobInfo{
		cancel:      context.CancelFunc(cancel),
		projectName: "testproject",
	})

	if !StopJob(jobID) {
		t.Error("StopJob returned false")
	}
	if !cancelled {
		t.Error("cancel function not called")
	}

	if StopJob("nonexistent") {
		t.Error("StopJob returned true for nonexistent job")
	}
}

func TestEnqueueJobDefaults(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "worker_enqueue_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	p := Project{
		Name:          "p1",
		RepoURL:       "https://repo",
		PrimaryBranch: "develop",
		BuildModel:    "m1",
	}
	SaveProject(&p)

	jobID, err := EnqueueJob(p, JobRequest{Prompt: "test"})
	if err != nil {
		t.Fatal(err)
	}

	job, err := readJob(filepath.Join(p.JobsPath("pending"), jobID+".yml"))
	if err != nil {
		t.Fatal(err)
	}

	if job.Request.Project != "p1" {
		t.Errorf("expected project p1, got %s", job.Request.Project)
	}
	if job.Request.RepoURL != "https://repo" {
		t.Errorf("expected repo https://repo, got %s", job.Request.RepoURL)
	}
	if job.Request.BranchParent != "develop" {
		t.Errorf("expected branch develop, got %s", job.Request.BranchParent)
	}
	if job.Request.Model != "m1" {
		t.Errorf("expected model m1, got %s", job.Request.Model)
	}
}

func TestMoveToCrash(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "worker_crash_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	p := Project{Name: "p1"}
	SaveProject(&p)

	jobID := "crashme"
	path := filepath.Join(p.JobsPath("pending"), jobID+".yml")
	job := Job{ID: jobID, Status: "pending"}
	data, _ := yaml.Marshal(&job)
	os.WriteFile(path, data, 0644)

	MoveToCrash(path, "some reason", p)

	// Check project is paused
	gp, _ := GetProject("p1")
	if !gp.Paused {
		t.Error("project should be paused after MoveToCrash")
	}

	// Check job status and error
	cj, _ := readJob(filepath.Join(p.JobsPath("crash"), jobID+".yml"))
	if cj.Status != "crash" || cj.Error != "some reason" {
		t.Errorf("unexpected job state: status=%s, error=%s", cj.Status, cj.Error)
	}
}
