package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleSettings(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-settings-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	req := httptest.NewRequest("GET", "/settings", nil)
	w := httptest.NewRecorder()

	HandleSettings(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Since we don't have a real project, it should still render the settings page
	if !strings.Contains(bodyStr, "Settings") {
		t.Errorf("Expected body to contain 'Settings'")
	}
}
