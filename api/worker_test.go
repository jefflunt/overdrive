package api

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScheduleAutoDocs(t *testing.T) {
	// Setup temporary projects directory
	tempDir, err := os.MkdirTemp("", "scheduler_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	// Create dummy scripts/work
	if err := os.MkdirAll("scripts", 0755); err != nil {
		t.Fatalf("Failed to create scripts dir: %v", err)
	}
	if err := os.WriteFile("scripts/work", []byte("#!/bin/sh\nexit 0"), 0755); err != nil {
		t.Fatalf("Failed to create dummy work script: %v", err)
	}

	// Create a project
	projectName := "test-project"
	project := Project{Name: projectName}
	if err := SaveProject(&project); err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}

	// Helper to enqueue a job
	enqueue := func(prompt string) {
		_, err := EnqueueJob(project, JobRequest{
			Project: projectName,
			Prompt:  prompt,
		})
		if err != nil {
			t.Fatalf("Failed to enqueue job: %v", err)
		}
	}

	// Enqueue 9 jobs
	for i := 1; i <= 9; i++ {
		enqueue("job")
	}

	// Run schedule 9 times to process these jobs
	for i := 1; i <= 9; i++ {
		schedule()
		// Wait for job to finish (move out of working)
		for {
			working, _ := filepath.Glob(project.JobsPath("working") + "/*.yml")
			if len(working) == 0 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Now enqueue the 10th job
	enqueue("job 10")
	schedule()

	// Wait for job to finish
	for {
		working, _ := filepath.Glob(project.JobsPath("working") + "/*.yml")
		if len(working) == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// After the 10th job is picked up, a docs job should have been enqueued
	pendingFiles, _ := filepath.Glob(project.JobsPath("pending") + "/*.yml")
	foundDocsJob := false
	for _, f := range pendingFiles {
		job, err := readJob(f)
		if err == nil && job.Request.Prompt == "/bdoc-update" {
			foundDocsJob = true
			break
		}
	}

	if !foundDocsJob {
		t.Errorf("Expected docs job to be enqueued after 10th job, but it wasn't")
	}
}

func TestScheduleAutoDocsPersisted(t *testing.T) {
	// Setup temporary projects directory
	tempDir, err := os.MkdirTemp("", "scheduler_test_persisted")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	// Create dummy scripts/work
	if err := os.MkdirAll("scripts", 0755); err != nil {
		t.Fatalf("Failed to create scripts dir: %v", err)
	}
	if err := os.WriteFile("scripts/work", []byte("#!/bin/sh\nexit 0"), 0755); err != nil {
		t.Fatalf("Failed to create dummy work script: %v", err)
	}

	// Create a project
	projectName := "test-project-persisted"
	project := Project{Name: projectName}
	if err := SaveProject(&project); err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}

	// Manually set the job count to 9 via file system
	// This simulates a restart where the count was preserved
	if err := writeJobCount(project, 9); err != nil {
		t.Fatalf("Failed to write job count: %v", err)
	}

	// Helper to enqueue a job
	enqueue := func(prompt string) {
		_, err := EnqueueJob(project, JobRequest{
			Project: projectName,
			Prompt:  prompt,
		})
		if err != nil {
			t.Fatalf("Failed to enqueue job: %v", err)
		}
	}

	// Enqueue the 10th job
	enqueue("job 10")
	schedule()

	// Check if docs job was enqueued
	pendingFiles, _ := filepath.Glob(project.JobsPath("pending") + "/*.yml")
	foundDocsJob := false
	for _, f := range pendingFiles {
		job, err := readJob(f)
		if err == nil && job.Request.Prompt == "/bdoc-update" {
			foundDocsJob = true
			break
		}
	}

	if !foundDocsJob {
		t.Errorf("Expected docs job to be enqueued after 10th job (starting from 9), but it wasn't")
	}

	// Check if count was reset to 0
	count, err := readJobCount(project)
	if err != nil {
		t.Fatalf("Failed to read job count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected job count to be reset to 0, but got %d", count)
	}
}
