package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHandleViewLogsExtra(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-view-logs-")
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
	jobID := "job1"
	os.MkdirAll(filepath.Join("projects", projectName, "logs", jobID), 0755)
	os.MkdirAll(filepath.Join("projects", projectName, "jobs", "done"), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "logs", jobID, "worker.log"), []byte("test log line"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "jobs", "done", jobID+".yml"), []byte("id: "+jobID+"\nstatus: done\nexit_code: 0\n"), 0644)

	req := httptest.NewRequest("GET", "/projects/"+projectName+"/jobs/logs/"+jobID, nil)
	w := httptest.NewRecorder()

	HandleViewLogs(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleProjectChatExtra(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-project-chat-")
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
	os.MkdirAll(filepath.Join("projects", projectName, "chats"), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)

	req := httptest.NewRequest("GET", "/projects/"+projectName+"/chat", nil)
	w := httptest.NewRecorder()

	HandleProjectChat(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleChatSendExtra(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-chat-send-")
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
	chatID := "chat1"
	os.MkdirAll(filepath.Join("projects", projectName, "chats"), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)
	os.WriteFile(filepath.Join("projects", projectName, "chats", "chat1.yml"), []byte("id: chat1\nproject: testproject\ntitle: test chat\n"), 0644)

	// Mock session
	sessionsMutex.Lock()
	sessions[projectName+":chat1"] = &ChatSession{
		ChatID:            chatID,
		ProjectName:       projectName,
		Port:              "12345",
		OpenCodeSessionID: "s1",
		LastActive:        time.Now(),
	}
	sessionsMutex.Unlock()
	defer func() {
		sessionsMutex.Lock()
		delete(sessions, projectName+":chat1")
		sessionsMutex.Unlock()
	}()

	req := httptest.NewRequest("POST", "/projects/"+projectName+"/chat/send", strings.NewReader("chat_id=chat1&content=hello"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	project, _ := GetProject(projectName)
	HandleChatSend(w, req, *project)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}
