package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func HandleListJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	projectName := getProjectName(r)
	if projectName == "" {
		// If no project specified, maybe redirect to projects list?
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	project, err := GetProject(projectName)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	allStatuses := []string{"pending", "working", "done", "crash", "no-op", "timeout", "cancelled"}
	statusQuery, hasStatus := r.URL.Query()["status"]
	var selectedStatuses []string
	if hasStatus {
		if statusQuery[0] != "" {
			selectedStatuses = strings.Split(statusQuery[0], ",")
		} else {
			selectedStatuses = []string{"__none__"}
		}
	}

	searchQuery := r.URL.Query().Get("q")
	initialPrompt := r.URL.Query().Get("prompt")

	// Pagination
	offset := 0
	if r.URL.Query().Get("offset") != "" {
		fmt.Sscanf(r.URL.Query().Get("offset"), "%d", &offset)
	}
	limit := 50
	if r.URL.Query().Get("limit") != "" {
		fmt.Sscanf(r.URL.Query().Get("limit"), "%d", &limit)
	}

	var allJobFiles []string
	dirsToLoad := allStatuses
	if len(selectedStatuses) > 0 {
		dirsToLoad = selectedStatuses
	}
	for _, d := range dirsToLoad {
		files, _ := filepath.Glob(filepath.Join(project.JobsPath(d), "*.yml"))
		allJobFiles = append(allJobFiles, files...)
	}

	// Sort files by name descending (reverse chronological order)
	sort.Slice(allJobFiles, func(i, j int) bool {
		return filepath.Base(allJobFiles[i]) > filepath.Base(allJobFiles[j])
	})

	if searchQuery != "" {
		var filtered []string
		query := strings.ToLower(searchQuery)
		for _, f := range allJobFiles {
			if job, err := readJob(f); err == nil && job != nil {
				if strings.Contains(strings.ToLower(job.RelatedCommit), query) ||
					strings.Contains(strings.ToLower(job.Request.Prompt), query) {
					filtered = append(filtered, f)
				}
			}
		}
		allJobFiles = filtered
	}

	totalJobs := len(allJobFiles)
	hasMore := false
	nextOffset := 0

	var jobs []Job
	if offset < totalJobs {
		end := offset + limit
		if end < totalJobs {
			hasMore = true
			nextOffset = end
		} else {
			end = totalJobs
		}

		for i := offset; i < end; i++ {
			if job, err := readJob(allJobFiles[i]); err == nil && job != nil {
				jobs = append(jobs, *job)
			}
		}
	} else {
		jobs = []Job{}
	}

	isHX := r.Header.Get("HX-Request") == "true"
	data := JobListPage{
		CurrentPath:      r.URL.Path,
		Project:          *project,
		Jobs:             jobs,
		Offset:           offset,
		Limit:            limit,
		NextOffset:       nextOffset,
		HasMore:          hasMore,
		TotalJobs:        totalJobs,
		AllStatuses:      allStatuses,
		SelectedStatuses: selectedStatuses,
		SearchQuery:      searchQuery,
		IsHX:             isHX,
		InitialPrompt:    initialPrompt,
	}

	if isHX && (r.Header.Get("HX-Target") == "jobs-container" || r.URL.Query().Get("only_rows") == "true") {
		log.Printf("[DEBUG] HTMX list request: project=%s, offset=%d, limit=%d, q=%s", projectName, offset, limit, searchQuery)
		tmpl, err := parseTemplate("api/templates/jobs.html")
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "jobs-list", data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	if isHX {
		tmpl, err := parseTemplate("api/templates/jobs.html")
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, "content", data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	tmpl, err := parseTemplate("api/templates/jobs.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func HandleTailLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	jobID := filepath.Base(r.URL.Path)
	if jobID == "tail" || jobID == "." || jobID == "/" {
		http.Error(w, "Invalid Job ID", http.StatusBadRequest)
		return
	}

	projectName := getProjectName(r)
	if projectName == "" {
		http.Error(w, "Project is required", http.StatusBadRequest)
		return
	}

	project, err := GetProject(projectName)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	logFile := filepath.Join(project.LogsPath(jobID), "worker.log")

	data, err := os.ReadFile(logFile)
	if err != nil {
		// If log file doesn't exist yet, just show empty
		data = []byte("Log file not found yet...")
	}

	// Find the job file
	var job *Job
	jobDirs := []string{"pending", "working", "done", "crash", "no-op", "timeout", "stopped", "undone", "cancelled"}
	for _, dir := range jobDirs {
		path := filepath.Join(project.JobsPath(dir), jobID+".yml")
		if j, err := readJob(path); err == nil {
			job = j
			break
		}
	}

	tmpl, err := parseTemplate("api/templates/tail_logs.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, LogPageData{
		CurrentPath: r.URL.Path,
		Project:     *project,

		JobID: jobID,
		Logs:  AnsiToHtml(string(data)),
		Job:   job,
	})
}

func HandleJobUpdates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	jobID := filepath.Base(r.URL.Path)
	if jobID == "updates" || jobID == "." || jobID == "/" {
		http.Error(w, "Invalid Job ID", http.StatusBadRequest)
		return
	}

	projectName := getProjectName(r)
	if projectName == "" {
		http.Error(w, "Project is required", http.StatusBadRequest)
		return
	}

	project, err := GetProject(projectName)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	var job *Job
	jobDirs := []string{"pending", "working", "done", "crash", "no-op", "timeout", "stopped", "undone", "cancelled"}
	for _, dir := range jobDirs {
		path := filepath.Join(project.JobsPath(dir), jobID+".yml")
		if j, err := readJob(path); err == nil {
			job = j
			break
		}
	}

	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Re-fetch project to ensure we have the latest state (like Paused status)
	// especially if the job just reached a terminal state.
	project, err = GetProject(projectName)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	tmpl, err := parseTemplate("api/templates/jobs.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Job     *Job
		Project *Project
	}{
		Job:     job,
		Project: project,
	}
	if err := tmpl.ExecuteTemplate(w, "job-updates", data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func HandleViewLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	jobID := filepath.Base(r.URL.Path)
	if jobID == "logs" || jobID == "." || jobID == "/" {
		http.Error(w, "Invalid Job ID", http.StatusBadRequest)
		return
	}

	projectName := getProjectName(r)
	if projectName == "" {
		http.Error(w, "Project is required", http.StatusBadRequest)
		return
	}

	project, err := GetProject(projectName)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	logFile := filepath.Join(project.LogsPath(jobID), "worker.log")

	data, err := os.ReadFile(logFile)
	if err != nil {
		http.Error(w, "Log file not found", http.StatusNotFound)
		return
	}

	// Find the job file to get exit code
	var job *Job
	jobDirs := []string{"done", "crash", "no-op", "working", "pending", "timeout", "stopped", "undone", "cancelled"}
	for _, dir := range jobDirs {
		path := filepath.Join(project.JobsPath(dir), jobID+".yml")
		if j, err := readJob(path); err == nil {
			job = j
			break
		}
	}

	exitCodeHtml := ""
	if job != nil && job.ExitCode != nil {
		errorHtml := ""
		if job.Error != "" {
			errorHtml = fmt.Sprintf(" <span style='color: #f14c4c;'>(%s)</span>", job.Error)
		}
		exitCodeHtml = fmt.Sprintf("<strong>Exit Code:</strong> %d%s", *job.ExitCode, errorHtml)
	}

	tmpl, err := parseTemplate("api/templates/view_logs.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, ViewLogsData{
		CurrentPath:  r.URL.Path,
		Project:      *project,
		JobID:        jobID,
		ExitCodeHtml: template.HTML(exitCodeHtml),
		Logs:         AnsiToHtml(string(data)),
		Job:          job,
	})
}

func HandlePartialLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	jobID := filepath.Base(r.URL.Path)
	if jobID == "logs-partial" || jobID == "." || jobID == "/" {
		http.Error(w, "Invalid Job ID", http.StatusBadRequest)
		return
	}

	projectName := getProjectName(r)
	if projectName == "" {
		http.Error(w, "Project is required", http.StatusBadRequest)
		return
	}

	project, err := GetProject(projectName)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	logFile := filepath.Join(project.LogsPath(jobID), "worker.log")

	data, err := os.ReadFile(logFile)
	if err != nil {
		// If log file doesn't exist yet, just show empty
		data = []byte("Log file not found yet...")
	} else if len(data) == 0 {
		data = []byte("Log file is empty...")
	}

	// Calculate MD5 of the content
	hash := md5.Sum(data)
	currentMD5 := hex.EncodeToString(hash[:])

	w.Header().Set("ETag", `"`+currentMD5+`"`)
	w.Header().Set("X-Log-MD5", currentMD5)
	w.Header().Set("Cache-Control", "no-cache")

	// Check if client has the same version
	if r.URL.Query().Get("md5sum") == currentMD5 || r.Header.Get("If-None-Match") == `"`+currentMD5+`"` {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Find the job file to get status
	var job *Job
	jobDirs := []string{"pending", "working", "done", "crash", "no-op", "timeout", "stopped", "undone", "cancelled"}
	for _, dir := range jobDirs {
		path := filepath.Join(project.JobsPath(dir), jobID+".yml")
		if j, err := readJob(path); err == nil {
			job = j
			break
		}
	}

	refreshAttr := ""
	if job != nil && (job.Status == "working" || job.Status == "pending") {
		refreshAttr = fmt.Sprintf("hx-get='/projects/%s/jobs/logs-partial/%s?md5sum=%s' hx-trigger='every 2s' hx-swap='outerHTML'", projectName, jobID, currentMD5)
	}

	innerID := fmt.Sprintf("logs-inner-%s", jobID)
	fmt.Fprintf(w, "<div id='%s' %s style='background-color: #1e1e1e; color: #d4d4d4; padding: 15px; font-family: monospace; max-height: 500px; overflow-y: auto;'><pre style='white-space: pre-wrap; word-wrap: break-word; margin: 0;'>%s</pre><script>(function(){ var el = document.getElementById('%s'); if(el) el.scrollTop = el.scrollHeight; })();</script></div>",
		innerID, refreshAttr, AnsiToHtml(string(data)), innerID)
}

func HandleStopJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet { // Allow GET for simple link clicks if needed, but POST is better. The prompt says "link" though.
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := filepath.Base(r.URL.Path)
	if jobID == "stop" || jobID == "." || jobID == "/" {
		http.Error(w, "Invalid Job ID", http.StatusBadRequest)
		return
	}

	projectName := getProjectName(r)
	if project, err := GetProject(projectName); err == nil {
		path := filepath.Join(project.JobsPath("working"), jobID+".yml")
		if job, err := readJob(path); err == nil && job != nil {
			if strings.HasPrefix(job.Request.Prompt, "/bdoc-update") {
				http.Error(w, "Docs jobs cannot be stopped", http.StatusForbidden)
				return
			}
		}
	}

	stopped := StopJob(jobID)
	if !stopped {
		http.Error(w, "Job not found or not running", http.StatusNotFound)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Stopping...")
		return
	}

	// Redirect back to jobs list
	http.Redirect(w, r, "/projects/"+projectName+"/jobs", http.StatusSeeOther)
}

func HandleCancelJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := filepath.Base(r.URL.Path)
	if jobID == "cancel" || jobID == "." || jobID == "/" {
		http.Error(w, "Invalid Job ID", http.StatusBadRequest)
		return
	}

	projectName := getProjectName(r)
	err := CancelJob(projectName, jobID)
	if err != nil {
		log.Printf("Failed to cancel job %s: %v", jobID, err)
		http.Error(w, "Failed to cancel job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Cancelled")
		return
	}

	// Redirect back to jobs list
	http.Redirect(w, r, "/projects/"+projectName+"/jobs", http.StatusSeeOther)
}

func HandleDeleteJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost { // Allow POST with _method=DELETE if needed, but HTMX uses DELETE
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := filepath.Base(r.URL.Path)
	projectName := getProjectName(r)
	if projectName == "" {
		http.Error(w, "Project is required", http.StatusBadRequest)
		return
	}

	project, err := GetProject(projectName)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Find the job file
	var jobPath string
	jobDirs := []string{"pending", "working", "done", "crash", "no-op", "timeout", "stopped", "undone", "cancelled"}
	for _, dir := range jobDirs {
		path := filepath.Join(project.JobsPath(dir), jobID+".yml")
		if _, err := os.Stat(path); err == nil {
			jobPath = path
			break
		}
	}

	if jobPath == "" {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Delete job file
	if err := os.Remove(jobPath); err != nil {
		http.Error(w, "Failed to delete job file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete logs directory
	logsPath := project.LogsPath(jobID)
	os.RemoveAll(logsPath)

	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Redirect back to jobs list
	http.Redirect(w, r, "/projects/"+projectName+"/jobs", http.StatusSeeOther)
}

func HandleSubmitJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Project == "" {
		req.Project = getProjectName(r)
	}
	if req.Project == "" {
		http.Error(w, "project is required", http.StatusBadRequest)
		return
	}

	project, err := GetProject(req.Project)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "prompt is required", http.StatusBadRequest)
		return
	}

	// Generate Job ID
	jobID, err := EnqueueJob(*project, req)
	if err != nil {
		log.Printf("Failed to enqueue job: %v", err)
		http.Error(w, "Failed to save job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": jobID})
}

func HandleRevertJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := filepath.Base(r.URL.Path)
	projectName := getProjectName(r)
	if projectName == "" {
		http.Error(w, "Project is required", http.StatusBadRequest)
		return
	}

	project, err := GetProject(projectName)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Find the job
	var job *Job
	path := filepath.Join(project.JobsPath("done"), jobID+".yml")
	if j, err := readJob(path); err == nil {
		job = j
	}

	if job == nil {
		http.Error(w, "Job not found or not in done status", http.StatusNotFound)
		return
	}

	commit := job.RelatedCommit
	if commit == "" {
		// If we don't have a captured commit, we'll pass the Job ID and let the worker
		// try to resolve it from the commit history.
		commit = jobID
	}

	// Submit a new revert job
	req := JobRequest{
		Project:      projectName,
		RepoURL:      project.RepoURL,
		BranchParent: project.PrimaryBranch,
		Prompt:       fmt.Sprintf("/bdoc-revert %s", commit),
		Model:        project.Build.LLM.Model,
	}
	req.CommitMsg = fmt.Sprintf("Revert to %s", commit)

	newJobID, err := EnqueueJob(*project, req)
	if err != nil {
		log.Printf("Failed to enqueue revert job: %v", err)
		http.Error(w, "Failed to save job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Revert job %s submitted", newJobID)
}

func HandleJobDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	jobID := filepath.Base(r.URL.Path)
	if jobID == "diff" || jobID == "." || jobID == "/" {
		http.Error(w, "Invalid Job ID", http.StatusBadRequest)
		return
	}

	projectName := getProjectName(r)
	if projectName == "" {
		http.Error(w, "Project is required", http.StatusBadRequest)
		return
	}

	project, err := GetProject(projectName)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Find the job
	var job *Job
	// Diffs are only relevant for done jobs usually, but let's check done first
	path := filepath.Join(project.JobsPath("done"), jobID+".yml")
	if j, err := readJob(path); err == nil {
		job = j
	} else {
		http.Error(w, "Job not found in done state", http.StatusNotFound)
		return
	}

	if job.RelatedCommit == "" {
		http.Error(w, "Job has no related commit", http.StatusBadRequest)
		return
	}

	// Read diff from file
	diffPath := filepath.Join(project.LogsPath(jobID), "diff")
	data, err := os.ReadFile(diffPath)
	if err != nil {
		// If diff file doesn't exist (older jobs), show a friendly message
		data = []byte("Diff not available for this job (logs cleared or job created before feature enabled).")
	}

	tmpl, err := parseTemplate("api/templates/diff.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, DiffPageData{
		CurrentPath: r.URL.Path,
		Project:     *project,

		JobID: jobID,
		Diff:  AnsiToHtml(string(data)),
		Job:   job,
	}); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}
