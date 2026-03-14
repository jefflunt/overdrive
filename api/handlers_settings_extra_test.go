package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleRebuildRestart(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-rebuild-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	// Create dummy script
	os.MkdirAll("scripts", 0755)
	os.WriteFile("scripts/rebuild-and-restart", []byte("#!/bin/sh\nexit 0\n"), 0755)

	req := httptest.NewRequest("POST", "/settings/rebuild-restart", nil)
	w := httptest.NewRecorder()

	HandleRebuildRestart(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "initiated") {
		t.Errorf("Expected body to contain 'initiated', got %s", string(body))
	}
}

func TestHandleRebuildRestartScheduler(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-rebuild-scheduler-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	// Create dummy script
	os.MkdirAll("scripts", 0755)
	os.WriteFile("scripts/rebuild-and-restart-scheduler", []byte("#!/bin/sh\nexit 0\n"), 0755)

	req := httptest.NewRequest("POST", "/settings/rebuild-restart-scheduler", nil)
	w := httptest.NewRecorder()

	HandleRebuildRestartScheduler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleHealthInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-health-info-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	req := httptest.NewRequest("GET", "/settings/health", nil)
	w := httptest.NewRecorder()

	HandleHealthInfo(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHandleHealthInfoHTMX(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-health-info-htmx-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	req := httptest.NewRequest("GET", "/settings/health", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	HandleHealthInfo(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "API Uptime") {
		t.Errorf("Expected body to contain 'API Uptime', got %s", string(body))
	}
}

func TestHandleSaveGlobalSettings(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-save-global-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	os.MkdirAll("projects", 0755)

	settings := GlobalSettings{
		MaxGlobalContainers:      5,
		MaxGlobalBuildContainers: 2,
		MaxGlobalChatContainers:  2,
		MaxGlobalCmdContainers:   2,
	}
	body, _ := json.Marshal(settings)

	req := httptest.NewRequest(http.MethodPost, "/settings/global/save", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	HandleSaveGlobalSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}

	// Verify global variables updated
	if MaxGlobalContainers != 5 {
		t.Errorf("Expected MaxGlobalContainers to be 5, got %d", MaxGlobalContainers)
	}

	// Verify file saved
	loaded, err := LoadGlobalSettings()
	if err != nil {
		t.Fatalf("Failed to load global settings: %v", err)
	}
	if loaded.MaxGlobalContainers != 5 {
		t.Errorf("Expected loaded MaxGlobalContainers to be 5, got %d", loaded.MaxGlobalContainers)
	}
}
