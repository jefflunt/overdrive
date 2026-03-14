package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleHelp(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api-test-help-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalWd, err := setupTestEnv(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	os.MkdirAll("help_docs", 0755)
	os.WriteFile("help_docs/JOBS.md", []byte("# Jobs Documentation\nContent here."), 0644)

	req := httptest.NewRequest("GET", "/help?doc=JOBS.md", nil)
	w := httptest.NewRecorder()

	HandleHelp(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected 200 OK, got %d: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "Jobs Documentation") {
		t.Errorf("Expected body to contain documentation title")
	}
}
