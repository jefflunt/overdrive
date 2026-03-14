package api

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestHandlePartialLogs(t *testing.T) {
	// Setup temporary project directory
	tmpDir, err := os.MkdirTemp("", "api-test-")
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
	jobID := "testjob"

	// Create project structure
	projectPath := filepath.Join("projects", projectName)
	os.MkdirAll(filepath.Join(projectPath, "jobs", "working"), 0755)
	os.MkdirAll(filepath.Join(projectPath, "logs", jobID), 0755)

	// Create project.yml
	projectYaml := []byte(fmt.Sprintf("name: %s\nrepo_url: https://github.com/example/repo.git\nssh_key: ssh-rsa ...\nprimary_branch: main\n", projectName))
	os.WriteFile(filepath.Join(projectPath, "project.yml"), projectYaml, 0644)

	// Create job.yml (needed to check status for polling)
	jobYaml := []byte(fmt.Sprintf("id: %s\nproject: %s\nstatus: working\nrequest:\n  prompt: test prompt\n", jobID, projectName))
	os.WriteFile(filepath.Join(projectPath, "jobs", "working", jobID+".yml"), jobYaml, 0644)

	// Create worker.log
	logContent := "Log Line 1\nLog Line 2\n"
	logPath := filepath.Join(projectPath, "logs", jobID, "worker.log")
	os.WriteFile(logPath, []byte(logContent), 0644)

	// 1. Initial Request (No MD5)
	req := httptest.NewRequest("GET", fmt.Sprintf("/projects/%s/jobs/logs-partial/%s", projectName, jobID), nil)
	w := httptest.NewRecorder()

	HandlePartialLogs(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "Log Line 1") {
		t.Errorf("Expected body to contain log content, got: %s", bodyStr)
	}

	// Extract MD5 from hx-get
	re := regexp.MustCompile(`md5sum=([a-f0-9]+)`)
	matches := re.FindStringSubmatch(bodyStr)
	if len(matches) < 2 {
		t.Fatalf("Could not find md5sum in response: %s", bodyStr)
	}
	md5sum := matches[1]

	// Verify MD5 calculation
	hash := md5.Sum([]byte(logContent))
	expectedMD5 := hex.EncodeToString(hash[:])
	if md5sum != expectedMD5 {
		t.Errorf("Expected MD5 %s, got %s", expectedMD5, md5sum)
	}

	// 2. Poll with Matching MD5 (Expect 304)
	req2 := httptest.NewRequest("GET", fmt.Sprintf("/projects/%s/jobs/logs-partial/%s?md5sum=%s", projectName, jobID, md5sum), nil)
	w2 := httptest.NewRecorder()

	HandlePartialLogs(w2, req2)

	resp2 := w2.Result()
	if resp2.StatusCode != http.StatusNotModified {
		t.Errorf("Expected 304 Not Modified, got %d", resp2.StatusCode)
	}

	body2, _ := io.ReadAll(resp2.Body)
	if len(body2) > 0 {
		t.Errorf("Expected empty body for 304, got %d bytes", len(body2))
	}

	// 3. Update Log and Poll with Old MD5 (Expect 200)
	newLogContent := logContent + "Log Line 3\n"
	os.WriteFile(logPath, []byte(newLogContent), 0644)

	req3 := httptest.NewRequest("GET", fmt.Sprintf("/projects/%s/jobs/logs-partial/%s?md5sum=%s", projectName, jobID, md5sum), nil)
	w3 := httptest.NewRecorder()

	HandlePartialLogs(w3, req3)

	resp3 := w3.Result()
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK after update, got %d", resp3.StatusCode)
	}

	body3, _ := io.ReadAll(resp3.Body)
	bodyStr3 := string(body3)
	if !strings.Contains(bodyStr3, "Log Line 3") {
		t.Errorf("Expected body to contain new log content")
	}

	// Verify new MD5 in response
	matches3 := re.FindStringSubmatch(bodyStr3)
	if len(matches3) < 2 {
		t.Fatalf("Could not find new md5sum in response")
	}
	newMD5 := matches3[1]
	// 4. Poll with ETag (Expect 304)
	req4 := httptest.NewRequest("GET", fmt.Sprintf("/projects/%s/jobs/logs-partial/%s", projectName, jobID), nil)
	req4.Header.Set("If-None-Match", `"`+newMD5+`"`)
	w4 := httptest.NewRecorder()

	HandlePartialLogs(w4, req4)

	resp4 := w4.Result()
	if resp4.StatusCode != http.StatusNotModified {
		t.Errorf("Expected 304 Not Modified with ETag, got %d", resp4.StatusCode)
	}

	body4, _ := io.ReadAll(resp4.Body)
	if len(body4) > 0 {
		t.Errorf("Expected empty body for 304, got %d bytes", len(body4))
	}
}

func TestHandleListJobs(t *testing.T) {
	// Setup temporary project directory
	tmpDir, err := os.MkdirTemp("", "api-test-list-jobs-")
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
	projectPath := filepath.Join("projects", projectName)
	os.MkdirAll(filepath.Join(projectPath, "jobs", "done"), 0755)

	// Create project.yml
	projectYaml := []byte(fmt.Sprintf("name: %s\nrepo_url: https://github.com/example/repo.git\nssh_key: ssh-rsa ...\nprimary_branch: main\n", projectName))
	os.WriteFile(filepath.Join(projectPath, "project.yml"), projectYaml, 0644)

	// Create a few job files
	for i := 1; i <= 3; i++ {
		jobID := fmt.Sprintf("job%d", i)
		jobYaml := []byte(fmt.Sprintf("id: %s\nproject: %s\nstatus: done\nrequest:\n  prompt: prompt %d\n", jobID, projectName, i))
		os.WriteFile(filepath.Join(projectPath, "jobs", "done", jobID+".yml"), jobYaml, 0644)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/projects/%s/jobs", projectName), nil)
	w := httptest.NewRecorder()

	HandleListJobs(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "prompt 1") || !strings.Contains(bodyStr, "prompt 2") || !strings.Contains(bodyStr, "prompt 3") {
		t.Errorf("Expected body to contain all job prompts")
	}
}
