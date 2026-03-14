package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConvertJiraIssuesToTodos(t *testing.T) {
	config := JiraConfig{ProjectKey: "PROJ"}
	issues := []JiraIssue{
		{
			ID:  "10001",
			Key: "PROJ-1",
			Fields: struct {
				Summary     string `json:"summary"`
				Description string `json:"description"`
				Status      struct {
					Name string `json:"name"`
				} `json:"status"`
				Issuetype struct {
					Name string `json:"name"`
				} `json:"issuetype"`
				Parent *struct {
					ID  string `json:"id"`
					Key string `json:"key"`
				} `json:"parent"`
			}{
				Summary: "Parent issue",
			},
		},
		{
			ID:  "10002",
			Key: "PROJ-2",
			Fields: struct {
				Summary     string `json:"summary"`
				Description string `json:"description"`
				Status      struct {
					Name string `json:"name"`
				} `json:"status"`
				Issuetype struct {
					Name string `json:"name"`
				} `json:"issuetype"`
				Parent *struct {
					ID  string `json:"id"`
					Key string `json:"key"`
				} `json:"parent"`
			}{
				Summary: "Child issue",
				Parent: &struct {
					ID  string `json:"id"`
					Key string `json:"key"`
				}{ID: "10001"},
			},
		},
	}

	todos := convertJiraIssuesToTodos(issues, config)
	if len(todos) != 1 {
		t.Fatalf("Expected 1 root todo, got %d", len(todos))
	}

	if todos[0].ID != "PROJ-1" {
		t.Errorf("Expected root ID PROJ-1, got %s", todos[0].ID)
	}

	if len(todos[0].Children) != 1 {
		t.Fatalf("Expected 1 child todo, got %d", len(todos[0].Children))
	}

	if todos[0].Children[0].ID != "PROJ-2" {
		t.Errorf("Expected child ID PROJ-2, got %s", todos[0].Children[0].ID)
	}
}

func TestFetchJiraIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := JiraSearchResponse{
			Issues: []JiraIssue{
				{
					Key: "PROJ-1",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := JiraConfig{
		Instance:   server.URL,
		ProjectKey: "PROJ",
		Email:      "test@example.com",
		APIToken:   "token",
	}

	// Clear cache
	jiraCacheMutex.Lock()
	delete(jiraCacheTime, config.ProjectKey)
	jiraCacheMutex.Unlock()

	todos, err := FetchJiraIssues(config)
	if err != nil {
		t.Fatal(err)
	}

	if len(todos) != 1 || todos[0].ID != "PROJ-1" {
		t.Errorf("Unexpected todos: %+v", todos)
	}
}
