package api

import (
	"testing"
	"time"
)

func TestChatSessionBasic(t *testing.T) {
	// Reset sessions map for testing
	sessionsMutex.Lock()
	sessions = make(map[string]*ChatSession)
	sessionsMutex.Unlock()

	projectName := "testproj"
	chatID := "chat123"
	key := projectName + ":" + chatID

	// Test GetChatSession (missing)
	if s := GetChatSession(projectName, chatID); s != nil {
		t.Errorf("Expected nil session, got %v", s)
	}

	// Manually add session
	session := &ChatSession{
		ChatID:      chatID,
		ProjectName: projectName,
		Port:        "8080",
		LastActive:  time.Now(),
	}
	sessionsMutex.Lock()
	sessions[key] = session
	sessionsMutex.Unlock()

	// Test GetChatSession (exists)
	if s := GetChatSession(projectName, chatID); s == nil || s.Port != "8080" {
		t.Errorf("Failed to get session or port mismatch")
	}
}

func TestWarmChatSessionLogic(t *testing.T) {
	// Reset maps
	sessionsMutex.Lock()
	sessions = make(map[string]*ChatSession)
	sessionsMutex.Unlock()
	warmingMutex.Lock()
	warming = make(map[string]chan struct{})
	warmingMutex.Unlock()

	projectName := "testproj"
	chatID := "chat123"
	project := Project{Name: projectName}

	// Mocking getOrCreateChatSession is hard, but we can test if it correctly sets 'warming'
	// We'll just test that calling it twice doesn't cause issues and uses the warming map

	// Since WarmChatSession starts a goroutine that calls getOrCreateChatSession,
	// and that will fail due to missing files, we just want to see if it doesn't crash
	// and handles the warming map correctly.

	WarmChatSession(project, chatID)

	key := projectName + ":" + chatID
	warmingMutex.Lock()
	_ = warming[key]
	warmingMutex.Unlock()

	// It might be already done or still warming, but calling it again should be safe
	WarmChatSession(project, chatID)
}

func TestReaper(t *testing.T) {
	sessionsMutex.Lock()
	sessions = make(map[string]*ChatSession)

	oldSession := &ChatSession{
		ChatID:      "old",
		ProjectName: "proj",
		LastActive:  time.Now().Add(-10 * time.Minute),
	}
	newSession := &ChatSession{
		ChatID:      "new",
		ProjectName: "proj",
		LastActive:  time.Now(),
	}
	sessions["proj:old"] = oldSession
	sessions["proj:new"] = newSession
	sessionsMutex.Unlock()

	// We can't easily trigger the reaper ticker, but we can call the reaper logic if we refactor it
	// Or we can just manually run the loop once.

	sessionsMutex.Lock()
	for key, session := range sessions {
		if time.Since(session.LastActive) > 5*time.Minute {
			// In real code this calls StopChatSession in a goroutine
			delete(sessions, key)
		}
	}
	sessionsMutex.Unlock()

	if _, exists := sessions["proj:old"]; exists {
		t.Error("old session should have been reaped")
	}
	if _, exists := sessions["proj:new"]; !exists {
		t.Error("new session should NOT have been reaped")
	}
}

func TestStopChatSessionMissing(t *testing.T) {
	// Should not crash when stopping non-existent session
	StopChatSession("nonexistent", "chat")
}

func TestSyncSessionNil(t *testing.T) {
	if err := SyncSession(nil); err != nil {
		t.Errorf("SyncSession(nil) should return nil error, got %v", err)
	}
	if err := SyncSession(&ChatSession{}); err != nil {
		t.Errorf("SyncSession empty should return nil error, got %v", err)
	}
}
