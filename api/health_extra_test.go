package api

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFormatUptime(t *testing.T) {
	d := 24*time.Hour + 2*time.Hour + 3*time.Minute
	s := FormatUptime(d)
	if s != "1d2h3m" {
		t.Errorf("Expected 1d2h3m, got %s", s)
	}
}

func TestSchedulerStartTime(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-health-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	err = RecordSchedulerStartTime()
	if err != nil {
		t.Fatal(err)
	}

	st, err := GetSchedulerStartTime()
	if err != nil {
		t.Fatal(err)
	}

	if time.Since(st) > time.Minute {
		t.Error("Recorded start time is too old")
	}
}

func TestGetRunningContainersInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-podman-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	// Create a mock podman
	os.Mkdir("bin", 0755)
	mockPodman := filepath.Join(tmpDir, "bin", "podman")
	os.WriteFile(mockPodman, []byte("#!/bin/sh\necho 'overdrive-chat-testproject-chat1|project=testproject,type=chat'"), 0755)

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", filepath.Join(tmpDir, "bin")+":"+oldPath)
	defer os.Setenv("PATH", oldPath)

	containers, err := GetRunningContainersInfo()
	if err != nil {
		t.Fatal(err)
	}

	if len(containers) != 1 {
		t.Fatalf("Expected 1 container, got %d", len(containers))
	}

	if containers[0].Name != "testproject-chat1" || containers[0].Project != "testproject" || containers[0].Type != "chat" {
		t.Errorf("Unexpected container info: %+v", containers[0])
	}
}
