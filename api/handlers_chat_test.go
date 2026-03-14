package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleChatCreate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-chat-create-")
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

	req := httptest.NewRequest("POST", fmt.Sprintf("/projects/%s/chat/create", projectName), nil)
	w := httptest.NewRecorder()

	project, _ := GetProject(projectName)
	HandleChatCreate(w, req, *project)

	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected 303 See Other, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if !strings.Contains(location, "/chat?id=") {
		t.Errorf("Expected redirect to chat, got %s", location)
	}
}

func TestHandleChatDelete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-chat-delete-")
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

	req := httptest.NewRequest("POST", "/projects/testproject/chat/delete/chat1", nil)
	w := httptest.NewRecorder()

	project, _ := GetProject(projectName)
	HandleChatDelete(w, req, *project, "chat1")

	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected 303 See Other, got %d", resp.StatusCode)
	}

	// Verify it's soft deleted
	chat, _ := GetChat(projectName, "chat1")
	if chat.DeletedAt == nil {
		t.Error("Expected chat to be deleted")
	}
}

func TestHandleChatList(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-chat-list-")
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
	os.WriteFile(filepath.Join("projects", projectName, "chats", "chat1.yml"), []byte("id: chat1\nproject: testproject\ntitle: chat number one\n"), 0644)

	req := httptest.NewRequest("GET", "/projects/testproject/chat/list", nil)
	w := httptest.NewRecorder()

	project, _ := GetProject(projectName)
	HandleChatList(w, req, *project)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "chat number one") {
		t.Errorf("Expected body to contain chat title, got: %s", string(body))
	}
}
