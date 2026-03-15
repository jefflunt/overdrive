package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func HandleListProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	projects, err := ListProjects()
	if err != nil {
		http.Error(w, "Failed to list projects: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := parseTemplate("api/templates/home.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := ProjectListPage{
		CurrentPath: r.URL.Path,
		Projects:    projects,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func HandleSaveProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var project Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if project.RepoURL == "" {
		http.Error(w, "Repo URL is required", http.StatusBadRequest)
		return
	}

	if project.SSHKey == "" {
		http.Error(w, "SSH Key is required", http.StatusBadRequest)
		return
	}

	if project.PrimaryBranch == "" {
		http.Error(w, "Working Branch is required", http.StatusBadRequest)
		return
	}

	if err := SaveProject(&project); err != nil {
		http.Error(w, "Failed to save project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(project)
}

func HandleGetProjectConfig(w http.ResponseWriter, r *http.Request) {
	name := getProjectName(r)
	if name == "" {
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}

	project, err := GetProject(name)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// For the UI, we want the content of the SSH key if it's currently a path
	if project.SSHKey != "" && filepath.IsAbs(project.SSHKey) {
		if content, err := os.ReadFile(project.SSHKey); err == nil {
			project.SSHKey = string(content)
		}
	}
	if project.SSHConfig != "" && filepath.IsAbs(project.SSHConfig) {
		if content, err := os.ReadFile(project.SSHConfig); err == nil {
			project.SSHConfig = string(content)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func HandleDeleteProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := filepath.Base(r.URL.Path)
	if err := DeleteProject(name); err != nil {
		http.Error(w, "Failed to delete project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HandleResumeProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := getProjectName(r)
	if name == "" {
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}

	if err := ResumeProject(name); err != nil {
		http.Error(w, "Failed to resume project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HandleRunProjectCmd(w http.ResponseWriter, r *http.Request) {
	name := getProjectName(r)
	if name == "" {
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	// /projects/{name}/cmds/{index}/run
	if len(parts) < 6 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	indexStr := parts[4]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	project, err := GetProject(name)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if index < 0 || index >= len(project.Cmds) {
		http.Error(w, "Command index out of range", http.StatusBadRequest)
		return
	}

	if err := CheckConcurrency(*project, "cmd"); err != nil {
		http.Error(w, "Concurrency limit reached: "+err.Error(), http.StatusTooManyRequests)
		return
	}

	cmd := project.Cmds[index]

	// Create a temporary job ID for this command
	jobID := "cmd-" + GenerateJobID()

	// Create log directory
	logDir := filepath.Join("projects", project.Name, "logs", jobID)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		http.Error(w, "Failed to create log directory", http.StatusInternalServerError)
		return
	}

	// Use Flusher to stream output
	flusher, _ := w.(http.Flusher)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)

	// Prepare command to run scripts/work
	cmdExec := exec.CommandContext(r.Context(), "./scripts/work")
	cmdExec.Env = os.Environ()

	// Set environment variables
	envMap := map[string]string{
		"JOB_ID":           jobID,
		"PROJECT_NAME":     project.Name,
		"REPO_URL":         project.RepoURL,
		"SSH_KEY":          project.SSHKey,
		"SSH_CONFIG":       project.SSHConfig,
		"WORKER_BRANCH":    project.PrimaryBranch,
		"BRANCH_PARENT":    project.PrimaryBranch,
		"COMMIT_MSG":       "Custom Command: " + cmd.Label,
		"PROMPT":           "Custom Command: " + cmd.Label,
		"CUSTOM_CMD":       cmd.Cmd,
		"MODEL":            project.Build.LLM.Model,
		"LLM_PROVIDER":     project.Build.LLM.Provider,
		"HARNESS_PROVIDER": project.Build.Harness.Provider,
		"AUTH_JSON":        project.GenerateAuthJSON(project.Build),
		"DEPENDENCIES":     project.Dependencies,
	}

	for k, v := range envMap {
		cmdExec.Env = append(cmdExec.Env, k+"="+v)
	}

	// Project custom environment variables
	var envNames []string
	for k, v := range project.EnvVars {
		cmdExec.Env = append(cmdExec.Env, k+"="+v)
		envNames = append(envNames, k)
	}
	if len(envNames) > 0 {
		cmdExec.Env = append(cmdExec.Env, "WORKER_ENV_NAMES="+strings.Join(envNames, ","))
	}

	// Use a writer that flushes after every write
	fw := flushingWriter{w: w, flusher: flusher}
	cmdExec.Stdout = fw
	cmdExec.Stderr = fw

	if err := cmdExec.Run(); err != nil {
		fmt.Fprintf(w, "\n\nError running command: %v\n", err)
	}

	if flusher != nil {
		flusher.Flush()
	}
}

type flushingWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func (fw flushingWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if fw.flusher != nil {
		fw.flusher.Flush()
	}
	return
}
