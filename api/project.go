package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type JiraConfig struct {
	Enabled      bool   `json:"enabled" yaml:"enabled"`
	Instance     string `json:"instance" yaml:"instance"`
	ProjectKey   string `json:"project_key" yaml:"project_key"`
	Email        string `json:"email" yaml:"email"`
	APIToken     string `json:"api_token" yaml:"api_token"`
	StatusPickup string `json:"status_pickup" yaml:"status_pickup"`
	StatusDone   string `json:"status_done" yaml:"status_done"`
}

type GitHubConfig struct {
	Repo         string `json:"repo" yaml:"repo"` // e.g. "owner/repo"
	Token        string `json:"token" yaml:"token"`
	StatusPickup string `json:"status_pickup" yaml:"status_pickup"`
	StatusDone   string `json:"status_done" yaml:"status_done"`
}

type ProjectCmd struct {
	Label string `json:"label" yaml:"label"`
	Cmd   string `json:"cmd" yaml:"cmd"`
}

type HarnessConfig struct {
	Provider string `json:"provider" yaml:"provider"`
}

type LLMConfig struct {
	Provider string `json:"provider" yaml:"provider"`
	Model    string `json:"model" yaml:"model"`
	APIKey   string `json:"api_key" yaml:"api_key"`
}

type UseCaseConfig struct {
	Harness HarnessConfig `json:"harness" yaml:"harness"`
	LLM     LLMConfig     `json:"llm" yaml:"llm"`
}

type ConcurrencyConfig struct {
	GlobalMax uint `json:"global_max" yaml:"global_max"`
	BuildMax  uint `json:"build_max" yaml:"build_max"`
	ChatMax   uint `json:"chat_max" yaml:"chat_max"`
	CmdMax    uint `json:"cmd_max" yaml:"cmd_max"`
}

type Project struct {
	Name          string            `json:"name" yaml:"name"`
	RepoURL       string            `json:"repo_url" yaml:"repo_url"`
	SSHKey        string            `json:"ssh_key" yaml:"ssh_key"`
	SSHConfig     string            `json:"ssh_config" yaml:"ssh_config"`
	PrimaryBranch string            `json:"primary_branch" yaml:"primary_branch"`
	BuildModel    string            `json:"build_model" yaml:"build_model"`
	ChatModel     string            `json:"chat_model" yaml:"chat_model"`
	Build         UseCaseConfig     `json:"build" yaml:"build"`
	Chat          UseCaseConfig     `json:"chat" yaml:"chat"`
	Dependencies  string            `json:"dependencies" yaml:"dependencies"`
	Cmds          []ProjectCmd      `json:"cmds" yaml:"cmds"`
	EnvVars       map[string]string `json:"env_vars" yaml:"env_vars"`
	TodoProvider  string            `json:"todo_provider" yaml:"todo_provider"`
	Paused        bool              `json:"paused" yaml:"paused"`
	Jira          JiraConfig        `json:"jira" yaml:"jira"`
	GitHub        GitHubConfig      `json:"github" yaml:"github"`
	Concurrency   ConcurrencyConfig `json:"concurrency" yaml:"concurrency"`
	LegacyModel   string            `json:"-" yaml:"model,omitempty"`
}

func (p Project) GenerateAuthJSON(useCase UseCaseConfig) string {
	if useCase.Harness.Provider != "opencode" {
		return "{}"
	}

	apiKey := useCase.LLM.APIKey
	if apiKey == "" {
		return "{}"
	}

	provider := useCase.LLM.Provider
	opencodeProvider := provider
	switch provider {
	case "google-gemini":
		opencodeProvider = "google"
	case "google-vertex":
		opencodeProvider = "google-vertex"
	case "opencode":
		opencodeProvider = "opencode"
	case "anthropic":
		opencodeProvider = "anthropic"
	}

	type AuthEntry struct {
		Type string `json:"type"`
		Key  string `json:"key"`
	}
	auth := make(map[string]AuthEntry)
	auth[opencodeProvider] = AuthEntry{
		Type: "api",
		Key:  apiKey,
	}

	data, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}

func (p Project) Path() string {
	return filepath.Join("projects", p.Name)
}

func (p Project) JobsPath(status string) string {
	return filepath.Join(p.Path(), "jobs", status)
}

func (p Project) LogsPath(jobID string) string {
	return filepath.Join(p.Path(), "logs", jobID)
}

func (p Project) ChatsPath() string {
	return filepath.Join(p.Path(), "chats")
}

func ListProjects() ([]Project, error) {
	var projects []Project
	entries, err := os.ReadDir("projects")
	if err != nil {
		if os.IsNotExist(err) {
			return projects, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			project, err := GetProject(entry.Name())
			if err == nil {
				projects = append(projects, *project)
			}
		}
	}
	return projects, nil
}

func GetProject(name string) (*Project, error) {
	path := filepath.Join("projects", name, "project.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var project Project
	if err := yaml.Unmarshal(data, &project); err != nil {
		return nil, err
	}
	project.Name = name

	// Migrate legacy model
	if project.BuildModel == "" && project.LegacyModel != "" {
		project.BuildModel = project.LegacyModel
		project.ChatModel = project.LegacyModel
	}

	// Initialize Build/Chat configs if empty
	if project.Build.Harness.Provider == "" {
		project.Build.Harness.Provider = "opencode"
	}
	if project.Build.LLM.Provider == "" {
		project.Build.LLM.Provider = "google-gemini"
	}
	if project.Build.LLM.Model == "" {
		project.Build.LLM.Model = project.BuildModel
		if project.Build.LLM.Model == "" {
			project.Build.LLM.Model = "google/gemini-3-flash-preview"
		}
	}

	if project.Chat.Harness.Provider == "" {
		project.Chat.Harness.Provider = "opencode"
	}
	if project.Chat.LLM.Provider == "" {
		project.Chat.LLM.Provider = "google-gemini"
	}
	if project.Chat.LLM.Model == "" {
		project.Chat.LLM.Model = project.ChatModel
		if project.Chat.LLM.Model == "" {
			project.Chat.LLM.Model = "google/gemini-3-flash-preview"
		}
	}

	// Migrate Jira enabled to TodoProvider
	if project.TodoProvider == "" {
		if project.Jira.Enabled {
			project.TodoProvider = "jira"
		} else {
			project.TodoProvider = "native"
		}
	}

	projectDir := filepath.Join("projects", name)

	// Check for ssh.key
	sshKeyPath := filepath.Join(projectDir, "ssh.key")
	if _, err := os.Stat(sshKeyPath); err == nil {
		if absPath, err := filepath.Abs(sshKeyPath); err == nil {
			project.SSHKey = absPath
		}
	}

	// Check for ssh_config
	sshConfigPath := filepath.Join(projectDir, "ssh_config")
	if _, err := os.Stat(sshConfigPath); err == nil {
		if absPath, err := filepath.Abs(sshConfigPath); err == nil {
			project.SSHConfig = absPath
		}
	}

	// Initialize Concurrency configs if empty
	if project.Concurrency.GlobalMax == 0 {
		project.Concurrency.GlobalMax = 1
	}
	if project.Concurrency.BuildMax == 0 {
		project.Concurrency.BuildMax = 1
	}
	if project.Concurrency.ChatMax == 0 {
		project.Concurrency.ChatMax = 1
	}
	if project.Concurrency.CmdMax == 0 {
		project.Concurrency.CmdMax = 1
	}

	return &project, nil
}

func SaveProject(project *Project) error {
	if project.Name == "" {
		// Derive name from RepoURL
		parts := strings.Split(project.RepoURL, "/")
		name := parts[len(parts)-1]
		name = strings.TrimSuffix(name, ".git")
		project.Name = name
	}

	path := filepath.Join("projects", project.Name)
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	// Create job dirs
	dirs := []string{"jobs/pending", "jobs/working", "jobs/done", "jobs/crash", "jobs/no-op", "jobs/timeout", "jobs/stopped", "jobs/undone", "jobs/cancelled", "logs", "chats"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(path, d), 0755); err != nil {
			return err
		}
	}

	// Write SSH key if it's provided as content (not a path)
	if project.SSHKey != "" && !filepath.IsAbs(project.SSHKey) {
		if err := os.WriteFile(filepath.Join(path, "ssh.key"), []byte(project.SSHKey), 0600); err != nil {
			return err
		}
	}

	// Write SSH config if it's provided as content (not a path)
	if project.SSHConfig != "" && !filepath.IsAbs(project.SSHConfig) {
		if err := os.WriteFile(filepath.Join(path, "ssh_config"), []byte(project.SSHConfig), 0644); err != nil {
			return err
		}
	}

	data, err := yaml.Marshal(project)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, "project.yml"), data, 0644)
}

func DeleteProject(name string) error {
	return os.RemoveAll(filepath.Join("projects", name))
}

func PauseProject(name string) error {
	p, err := GetProject(name)
	if err != nil {
		return err
	}
	p.Paused = true
	return SaveProject(p)
}

func ResumeProject(name string) error {
	p, err := GetProject(name)
	if err != nil {
		return err
	}
	p.Paused = false
	return SaveProject(p)
}
