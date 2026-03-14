package api

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestChatStorage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-chat-storage-")
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
	os.MkdirAll(filepath.Join("projects", projectName), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)

	chatID := GenerateChatID()
	chat := &Chat{
		ID:        chatID,
		Project:   projectName,
		Title:     "Test Chat",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello", Timestamp: time.Now()},
		},
	}

	// Save
	if err := SaveChat(chat); err != nil {
		t.Fatalf("SaveChat failed: %v", err)
	}

	// Get
	loaded, err := GetChat(projectName, chatID)
	if err != nil {
		t.Fatalf("GetChat failed: %v", err)
	}
	if loaded.Title != "Test Chat" {
		t.Errorf("Expected title Test Chat, got %s", loaded.Title)
	}
	if len(loaded.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(loaded.Messages))
	}

	// List
	chats, err := ListChats(projectName)
	if err != nil {
		t.Fatalf("ListChats failed: %v", err)
	}
	if len(chats) != 1 {
		t.Errorf("Expected 1 chat, got %d", len(chats))
	}

	// Delete (soft)
	if err := DeleteChat(projectName, chatID); err != nil {
		t.Fatalf("DeleteChat failed: %v", err)
	}

	chats, _ = ListChats(projectName)
	if len(chats) != 0 {
		t.Errorf("Expected 0 active chats, got %d", len(chats))
	}

	deletedChats, _ := ListDeletedChats(projectName)
	if len(deletedChats) != 1 {
		t.Errorf("Expected 1 deleted chat, got %d", len(deletedChats))
	}

	// Restore
	if err := RestoreChat(projectName, chatID); err != nil {
		t.Fatalf("RestoreChat failed: %v", err)
	}
	chats, _ = ListChats(projectName)
	if len(chats) != 1 {
		t.Errorf("Expected 1 chat after restore, got %d", len(chats))
	}

	// Permanently delete
	if err := PermanentlyDeleteChat(projectName, chatID); err != nil {
		t.Fatalf("PermanentlyDeleteChat failed: %v", err)
	}
	chats, _ = ListChats(projectName)
	if len(chats) != 0 {
		t.Errorf("Expected 0 chats after permanent delete, got %d", len(chats))
	}
}

func TestCleanupOldDeletedChats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-chat-cleanup-")
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
	os.MkdirAll(filepath.Join("projects", projectName), 0755)
	os.WriteFile(filepath.Join("projects", projectName, "project.yml"), []byte("name: "+projectName+"\n"), 0644)

	oldDate := time.Now().AddDate(0, 0, -70)
	chatID := "oldchat"
	chat := &Chat{
		ID:        chatID,
		Project:   projectName,
		DeletedAt: &oldDate,
		UpdatedAt: oldDate,
	}
	SaveChat(chat)

	if err := CleanupOldDeletedChats(projectName); err != nil {
		t.Fatalf("CleanupOldDeletedChats failed: %v", err)
	}

	deletedChats, _ := ListDeletedChats(projectName)
	if len(deletedChats) != 0 {
		t.Errorf("Expected 0 deleted chats after cleanup, got %d", len(deletedChats))
	}
}
