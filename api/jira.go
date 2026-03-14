package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type JiraIssue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
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
	} `json:"fields"`
}

type JiraSearchResponse struct {
	Issues []JiraIssue `json:"issues"`
}

var (
	jiraCache      = make(map[string][]Todo)
	jiraCacheMutex sync.Mutex
	jiraCacheTime  = make(map[string]time.Time)
)

const cacheDuration = 1 * time.Minute

func FetchJiraIssues(config JiraConfig) ([]Todo, error) {
	jiraCacheMutex.Lock()
	if t, ok := jiraCacheTime[config.ProjectKey]; ok && time.Since(t) < cacheDuration {
		todos := jiraCache[config.ProjectKey]
		jiraCacheMutex.Unlock()
		return todos, nil
	}
	jiraCacheMutex.Unlock()

	// Load API token from env if it looks like an env var
	apiToken := config.APIToken
	if strings.HasPrefix(apiToken, "$") {
		apiToken = os.Getenv(strings.TrimPrefix(apiToken, "$"))
	}

	url := fmt.Sprintf("%s/rest/api/3/search?jql=project=%s&maxResults=100", config.Instance, config.ProjectKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(config.Email, apiToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API error: %d %s", resp.StatusCode, string(body))
	}

	var searchResp JiraSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	todos := convertJiraIssuesToTodos(searchResp.Issues, config)

	jiraCacheMutex.Lock()
	jiraCache[config.ProjectKey] = todos
	jiraCacheTime[config.ProjectKey] = time.Now()
	jiraCacheMutex.Unlock()

	return todos, nil
}

func convertJiraIssuesToTodos(issues []JiraIssue, config JiraConfig) []Todo {
	// 1. Create all base Todo objects
	issueMap := make(map[string]*Todo)
	for _, issue := range issues {
		todo := &Todo{
			ID:          issue.Key,
			ProjectID:   config.ProjectKey,
			Title:       issue.Fields.Summary,
			Description: issue.Fields.Description,
			Status:      issue.Fields.Status.Name,
			CreatedAt:   time.Now().Unix(),
		}
		issueMap[issue.ID] = todo
		issueMap[issue.Key] = todo
	}

	// 2. Build parent-child mapping
	parentToChildren := make(map[string][]*Todo)
	for _, issue := range issues {
		if issue.Fields.Parent != nil {
			parentID := issue.Fields.Parent.ID
			parentKey := issue.Fields.Parent.Key
			
			parent, ok := issueMap[parentID]
			if !ok {
				parent, ok = issueMap[parentKey]
			}
			
			if ok {
				child := issueMap[issue.ID]
				child.ParentID = parent.ID
				parentToChildren[parent.ID] = append(parentToChildren[parent.ID], child)
			}
		}
	}

	// 3. Recursive function to assemble the final tree (with value copies)
	var assemble func(*Todo) Todo
	assemble = func(t *Todo) Todo {
		result := *t
		children := parentToChildren[t.ID]
		for _, child := range children {
			result.Children = append(result.Children, assemble(child))
		}
		return result
	}

	// 4. Collect roots and assemble them
	var rootTodos []Todo
	seenRoots := make(map[string]bool)
	for _, issue := range issues {
		todo := issueMap[issue.ID]
		isRoot := true
		if issue.Fields.Parent != nil {
			if _, ok := issueMap[issue.Fields.Parent.ID]; ok {
				isRoot = false
			} else if _, ok := issueMap[issue.Fields.Parent.Key]; ok {
				isRoot = false
			}
		}
		
		if isRoot && !seenRoots[todo.ID] {
			rootTodos = append(rootTodos, assemble(todo))
			seenRoots[todo.ID] = true
		}
	}

	return rootTodos
}

func UpdateJiraIssueStatus(config JiraConfig, issueKey string, statusName string) error {
	// Load API token from env if it looks like an env var
	apiToken := config.APIToken
	if strings.HasPrefix(apiToken, "$") {
		apiToken = os.Getenv(strings.TrimPrefix(apiToken, "$"))
	}

	// First, find the transition ID for the status name
	transitionID, err := findTransitionID(config, issueKey, statusName, apiToken)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", config.Instance, issueKey)
	payload := map[string]interface{}{
		"transition": map[string]string{
			"id": transitionID,
		},
	}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	req.SetBasicAuth(config.Email, apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Jira transition error: %d %s", resp.StatusCode, string(body))
	}

	// Invalidate cache
	jiraCacheMutex.Lock()
	delete(jiraCacheTime, config.ProjectKey)
	jiraCacheMutex.Unlock()

	return nil
}

func findTransitionID(config JiraConfig, issueKey string, statusName string, apiToken string) (string, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", config.Instance, issueKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(config.Email, apiToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Jira API error fetching transitions: %d %s", resp.StatusCode, string(body))
	}

	var result struct {
		Transitions []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			To   struct {
				Name string `json:"name"`
			} `json:"to"`
		} `json:"transitions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	for _, t := range result.Transitions {
		if strings.EqualFold(t.Name, statusName) || strings.EqualFold(t.To.Name, statusName) {
			return t.ID, nil
		}
	}

	return "", fmt.Errorf("transition to status '%s' not found", statusName)
}
