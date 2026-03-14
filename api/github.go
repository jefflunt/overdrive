package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type GitHubIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

var (
	githubCache      = make(map[string][]Todo)
	githubCacheMutex sync.Mutex
	githubCacheTime  = make(map[string]time.Time)
)

func FetchGitHubIssues(config GitHubConfig) ([]Todo, error) {
	if config.Repo == "" {
		return []Todo{}, nil
	}
	githubCacheMutex.Lock()
	if t, ok := githubCacheTime[config.Repo]; ok && time.Since(t) < cacheDuration {
		todos := githubCache[config.Repo]
		githubCacheMutex.Unlock()
		return todos, nil
	}
	githubCacheMutex.Unlock()

	// Load token from env if it looks like an env var
	token := config.Token
	if strings.HasPrefix(token, "$") {
		token = os.Getenv(strings.TrimPrefix(token, "$"))
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/issues?state=open&per_page=100", config.Repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d %s", resp.StatusCode, string(body))
	}

	var issues []GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	var todos []Todo
	for _, issue := range issues {
		// GitHub issues API also returns Pull Requests. PRs have a 'pull_request' field.
		// We want to skip PRs if they are not intended to be todos.
		// However, for simplicity, we'll just include everything that comes back.

		status := issue.State
		for _, l := range issue.Labels {
			if strings.EqualFold(l.Name, config.StatusPickup) || strings.EqualFold(l.Name, config.StatusDone) {
				status = l.Name
				break
			}
		}

		todos = append(todos, Todo{
			ID:          strconv.Itoa(issue.Number),
			ProjectID:   config.Repo,
			Title:       issue.Title,
			Description: issue.Body,
			Status:      status,
			CreatedAt:   time.Now().Unix(),
		})
	}

	githubCacheMutex.Lock()
	githubCache[config.Repo] = todos
	githubCacheTime[config.Repo] = time.Now()
	githubCacheMutex.Unlock()

	return todos, nil
}

func UpdateGitHubIssueStatus(config GitHubConfig, issueNumber string, statusName string) error {
	if config.Repo == "" || issueNumber == "" {
		return nil
	}
	// Load token from env if it looks like an env var
	token := config.Token
	if strings.HasPrefix(token, "$") {
		token = os.Getenv(strings.TrimPrefix(token, "$"))
	}

	if statusName == "" {
		return nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s", config.Repo, issueNumber)

	payload := make(map[string]interface{})

	if strings.EqualFold(statusName, "closed") || strings.EqualFold(statusName, config.StatusDone) {
		payload["state"] = "closed"
	} else if strings.EqualFold(statusName, "open") {
		payload["state"] = "open"
	}

	// If it's a label-based status, we might want to add/remove labels,
	// but for now let's keep it simple and just handle state if it's "closed".
	// A more robust implementation would handle labels.

	if len(payload) == 0 {
		return nil
	}

	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error updating issue: %d %s", resp.StatusCode, string(body))
	}

	// Invalidate cache
	githubCacheMutex.Lock()
	delete(githubCacheTime, config.Repo)
	githubCacheMutex.Unlock()

	return nil
}
