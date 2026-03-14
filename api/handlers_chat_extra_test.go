package api

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHandleChatRestoreExtra(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-chat-restore-")
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

	chat := Chat{
		ID:        "chat1",
		Project:   projectName,
		Title:     "deleted chat",
		DeletedAt: func() *time.Time { t := time.Now(); return &t }(),
	}
	SaveChat(&chat)

	req := httptest.NewRequest("POST", "/projects/testproject/chat/restore/chat1", nil)
	w := httptest.NewRecorder()

	project, _ := GetProject(projectName)
	HandleChatRestore(w, req, *project, "chat1")

	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected 303 See Other, got %d", resp.StatusCode)
	}

	// Verify it's restored
	restored, _ := GetChat(projectName, "chat1")
	if restored.DeletedAt != nil {
		t.Error("Expected chat to be restored (DeletedAt should be nil)")
	}
}

func TestHandleChatMessagesExtra(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-chat-messages-")
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
	os.WriteFile(filepath.Join("projects", projectName, "chats", "chat1.yml"), []byte("id: chat1\nproject: testproject\ntitle: test chat\nmessages: [{role: user, content: hello}]\n"), 0644)

	req := httptest.NewRequest("GET", "/projects/testproject/chat/messages/chat1", nil)
	w := httptest.NewRecorder()

	project, _ := GetProject(projectName)
	HandleChatMessages(w, req, *project, "chat1", nil)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleChatSyncExtra(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-chat-sync-")
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
	os.WriteFile(filepath.Join("projects", projectName, "chats", "chat1.yml"), []byte("id: chat1\nproject: testproject\ntitle: test chat\n"), 0644)

	// Mock session to avoid calling startChatServer
	sessionsMutex.Lock()
	sessions[projectName+":chat1"] = &ChatSession{
		ChatID:      "chat1",
		ProjectName: projectName,
		Port:        "12345",
		LastActive:  time.Now(),
	}
	sessionsMutex.Unlock()
	defer func() {
		sessionsMutex.Lock()
		delete(sessions, projectName+":chat1")
		sessionsMutex.Unlock()
	}()

	req := httptest.NewRequest("GET", "/projects/testproject/chat/sync/chat1", nil)
	w := httptest.NewRecorder()

	project, _ := GetProject(projectName)
	HandleChatSync(w, req, *project, "chat1")

	resp := w.Result()
	// It will try to connect to localhost:12345 and fail, but it should still return 200 with an error template or something
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleChatDeletePermanent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-chat-delete-perm-")
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
	os.WriteFile(filepath.Join("projects", projectName, "chats", "chat1.yml"), []byte("id: chat1\nproject: testproject\ntitle: test chat\n"), 0644)

	req := httptest.NewRequest("POST", "/projects/testproject/chat/delete/chat1?permanent=true", nil)
	w := httptest.NewRecorder()

	project, _ := GetProject(projectName)
	HandleChatDelete(w, req, *project, "chat1")

	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected 303 See Other, got %d", resp.StatusCode)
	}

	// Verify it's gone
	if _, err := os.Stat(filepath.Join("projects", projectName, "chats", "chat1.yml")); !os.IsNotExist(err) {
		t.Error("Expected chat file to be deleted permanently")
	}
}

func TestHandleChatBuild(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-chat-build-")
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
	project := Project{Name: projectName}

	req := httptest.NewRequest("POST", "/projects/testproject/chat/build", strings.NewReader("content=test build"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	HandleChatBuild(w, req, project)

	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected 303 See Other, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if !strings.Contains(location, "/projects/testproject/jobs?prompt=test+build") {
		t.Errorf("Expected redirect to jobs with prompt, got %s", location)
	}
}

func TestHandleChatProxy(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "proxy_test")
	defer os.RemoveAll(tmpDir)
	origWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	p := Project{Name: "testproj"}
	if err := SaveProject(&p); err != nil {
		t.Fatal(err)
	}

	backendCalled := false
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/global/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		backendCalled = true
		if r.URL.Path != "/foo" {
			t.Errorf("Expected path /foo, got %s", r.URL.Path)
		}
		w.Write([]byte("backend ok"))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	_, port, _ := net.SplitHostPort(u.Host)

	chatID := "chat123"
	sessionsMutex.Lock()
	sessions["testproj:chat123"] = &ChatSession{
		ChatID:      chatID,
		ProjectName: "testproj",
		Port:        port,
		LastActive:  time.Now(),
	}
	sessionsMutex.Unlock()

	req := httptest.NewRequest("GET", "/projects/testproj/chat/proxy/chat123/foo", nil)
	w := httptest.NewRecorder()
	HandleChatProxy(w, req, p, chatID)

	if !backendCalled {
		t.Error("backend was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "backend ok" {
		t.Errorf("expected 'backend ok', got %q", w.Body.String())
	}
}

func TestHandleChatRename(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "rename_test")
	defer os.RemoveAll(tmpDir)
	origWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	p := Project{Name: "testproj"}
	if err := SaveProject(&p); err != nil {
		t.Fatal(err)
	}
	chatID := "chat123"
	chat := Chat{ID: chatID, Project: "testproj", Title: "Old Title"}
	if err := SaveChat(&chat); err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("title", "New Title")
	req := httptest.NewRequest("POST", "/projects/testproj/chat/rename/chat123", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	HandleChatRename(w, req, p, chatID)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}

	updatedChat, _ := GetChat("testproj", chatID)
	if updatedChat.Title != "New Title" {
		t.Errorf("expected 'New Title', got %q", updatedChat.Title)
	}
}
