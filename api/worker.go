package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

var MaxProjectConcurrentJobs = 1
var MaxGlobalContainers = 10
var MaxGlobalBuildContainers = 5
var MaxGlobalChatContainers = 5
var MaxGlobalCmdContainers = 5

type jobInfo struct {
	cancel      context.CancelFunc
	projectName string
}

var activeContexts sync.Map // map[string]jobInfo

func CleanupZombieJobs() {
	projects, err := ListProjects()
	if err != nil {
		log.Printf("Error listing projects for cleanup: %v", err)
		return
	}

	for _, p := range projects {
		var allFiles []string
		for _, status := range []string{"working"} {
			files, err := filepath.Glob(p.JobsPath(status) + "/*.yml")
			if err == nil {
				allFiles = append(allFiles, files...)
			}
		}

		for _, f := range allFiles {
			log.Printf("Found zombie job %s in project %s, moving to crash...", f, p.Name)
			MoveToCrash(f, "Job interrupted by server restart", p)
		}
	}
}

func StartScheduler() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			schedule()
		}
	}()
}

func EnqueueJob(project Project, req JobRequest) (string, error) {
	// Defaults
	if req.Project == "" {
		req.Project = project.Name
	}
	if req.RepoURL == "" {
		req.RepoURL = project.RepoURL
	}
	if req.BranchParent == "" {
		if project.PrimaryBranch != "" {
			req.BranchParent = project.PrimaryBranch
		} else {
			req.BranchParent = "main"
		}
	}
	if req.Model == "" {
		if project.Build.LLM.Model != "" {
			req.Model = project.Build.LLM.Model
		} else if project.BuildModel != "" {
			req.Model = project.BuildModel
		} else {
			req.Model = "google/gemini-3-flash-preview"
		}
	}

	jobID := GenerateJobID()

	if req.CommitMsg == "" {
		if strings.HasPrefix(req.Prompt, "/bdoc-quick") {
			cleaned := CleanPromptForCommit(req.Prompt)
			if cleaned != "" {
				req.CommitMsg = cleaned
			} else {
				req.CommitMsg = jobID
			}
		} else {
			req.CommitMsg = jobID
		}
	}

	job := Job{
		ID:          jobID,
		Project:     project.Name,
		Status:      "pending",
		Repo:        req.RepoURL,
		CreatedAt:   time.Now(),
		Request:     req,
		ReferenceID: req.ReferenceID,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&job)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job: %w", err)
	}

	// Write to project's jobs/pending
	filename := filepath.Join(project.JobsPath("pending"), jobID+".yml")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write job file: %w", err)
	}

	return jobID, nil
}

func CheckConcurrency(project Project, containerType string) error {
	allCounts := GetAllRunningContainerCounts()

	// For each project, check how many jobs are in 'working' status
	// to avoid starting new jobs before the previous ones have registered as containers.
	workingJobs := make(map[string]int)
	projects, err := ListProjects()
	if err == nil {
		for _, p := range projects {
			files, _ := filepath.Glob(p.JobsPath("working") + "/*.yml")
			workingJobs[p.Name] = len(files)
		}
	}

	// Check global system limit
	totalGlobal := 0
	buildGlobal := 0
	chatGlobal := 0
	cmdGlobal := 0
	for _, c := range allCounts {
		totalGlobal += c.Total()
		buildGlobal += c.Build
		chatGlobal += c.Chat
		cmdGlobal += c.Cmd
		// If we see containers, we assume those jobs are accounted for.
	}

	// Also add working jobs that haven't become containers yet
	// (Note: this might over-count slightly if a container just started, but that's safe for concurrency)
	for _, count := range workingJobs {
		// Only count as buildGlobal if there isn't already a container for that project's build
		// Actually, let's just add it, it's safer for strict concurrency.
		totalGlobal += count
		buildGlobal += count
	}

	if totalGlobal >= MaxGlobalContainers {
		return fmt.Errorf("system-wide concurrency limit reached (%d/%d)", totalGlobal, MaxGlobalContainers)
	}

	switch containerType {
	case "build":
		if buildGlobal >= MaxGlobalBuildContainers {
			return fmt.Errorf("global build concurrency limit reached (%d/%d)", buildGlobal, MaxGlobalBuildContainers)
		}
	case "chat":
		if chatGlobal >= MaxGlobalChatContainers {
			return fmt.Errorf("global chat concurrency limit reached (%d/%d)", chatGlobal, MaxGlobalChatContainers)
		}
	case "cmd":
		if cmdGlobal >= MaxGlobalCmdContainers {
			return fmt.Errorf("global cmd concurrency limit reached (%d/%d)", cmdGlobal, MaxGlobalCmdContainers)
		}
	}

	c := allCounts[project.Name]
	total := c.Total() + workingJobs[project.Name]

	if uint(total) >= project.Concurrency.GlobalMax {
		return fmt.Errorf("project global concurrency limit reached (%d/%d)", total, project.Concurrency.GlobalMax)
	}

	switch containerType {
	case "build":
		// Per-project build constraint: no more than 1 build container per project.
		// This constraint is not configurable.
		if c.Build+workingJobs[project.Name] >= 1 {
			return fmt.Errorf("build concurrency limit reached (%d/1)", c.Build+workingJobs[project.Name])
		}
	case "chat":
		if uint(c.Chat) >= project.Concurrency.ChatMax {
			return fmt.Errorf("chat concurrency limit reached (%d/%d)", c.Chat, project.Concurrency.ChatMax)
		}
	case "cmd":
		if uint(c.Cmd) >= project.Concurrency.CmdMax {
			return fmt.Errorf("cmd concurrency limit reached (%d/%d)", c.Cmd, project.Concurrency.CmdMax)
		}
	}

	return nil
}

type ContainerCounts struct {
	Build int
	Chat  int
	Cmd   int
}

func (c ContainerCounts) Total() int {
	return c.Build + c.Chat + c.Cmd
}

func GetAllRunningContainerCounts() map[string]ContainerCounts {
	counts := make(map[string]ContainerCounts)
	cmdExec := exec.Command("podman", "ps", "--format", "{{.Labels}}")
	out, err := cmdExec.Output()
	if err != nil {
		log.Printf("Error counting all containers: %v", err)
		return counts
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Labels are usually comma-separated in --format {{.Labels}}
		// E.g. project=foo,type=build
		project := ""
		cType := ""

		parts := strings.Split(line, ",")
		for _, part := range parts {
			kv := strings.Split(part, "=")
			if len(kv) == 2 {
				if kv[0] == "project" {
					project = kv[1]
				} else if kv[0] == "type" {
					cType = kv[1]
				}
			}
		}

		if project != "" {
			c := counts[project]
			switch cType {
			case "build":
				c.Build++
			case "chat":
				c.Chat++
			case "cmd":
				c.Cmd++
			}
			counts[project] = c
		}
	}
	return counts
}

func GetRunningContainerCounts(project Project) (build, chat, cmd int) {
	allCounts := GetAllRunningContainerCounts()
	c := allCounts[project.Name]
	return c.Build, c.Chat, c.Cmd
}

func schedule() {
	projects, err := ListProjects()
	if err != nil {
		log.Printf("Error listing projects for scheduling: %v", err)
		return
	}

	allCounts := GetAllRunningContainerCounts()

	// For each project, check how many jobs are in 'working' status
	// to avoid starting new jobs before the previous ones have registered as containers.
	workingJobsMap := make(map[string]int)
	for _, p := range projects {
		files, _ := filepath.Glob(p.JobsPath("working") + "/*.yml")
		workingJobsMap[p.Name] = len(files)
	}

	// Calculate total working containers across all projects for global limit
	totalWorkingGlobal := 0
	totalWorkingBuild := 0
	for _, c := range allCounts {
		totalWorkingGlobal += c.Total()
		totalWorkingBuild += c.Build
	}

	// Add working jobs to global counts
	for _, count := range workingJobsMap {
		totalWorkingGlobal += count
		totalWorkingBuild += count
	}

	if totalWorkingGlobal >= MaxGlobalContainers {
		return
	}
	if totalWorkingBuild >= MaxGlobalBuildContainers {
		return
	}

	// For each project, check if it can run a job
	for _, p := range projects {
		if p.Paused {
			continue
		}

		c := allCounts[p.Name]
		workingJobs := workingJobsMap[p.Name]
		totalRunning := c.Total() + workingJobs

		// Global project max
		if uint(totalRunning) >= p.Concurrency.GlobalMax {
			continue
		}
		// Per-project build constraint: no more than 1 build container per project.
		// This constraint is not configurable.
		if c.Build+workingJobs >= 1 {
			continue
		}

		// Find pending jobs for this project
		files, _ := filepath.Glob(p.JobsPath("pending") + "/*.yml")
		if len(files) == 0 {
			continue
		}

		// Sort by ID (chronological)
		sort.Strings(files)

		// Pick the oldest job
		selectedJobPath := files[0]
		filename := filepath.Base(selectedJobPath)

		// Read job to check if it's a docs job
		job, err := readJob(selectedJobPath)
		isDocsJob := false
		if err == nil && job != nil {
			isDocsJob = strings.HasPrefix(job.Request.Prompt, "/bdoc-update")
		}

		// Move to working
		workingPath := filepath.Join(p.JobsPath("working"), filename)
		if err := os.Rename(selectedJobPath, workingPath); err != nil {
			log.Printf("Failed to move job %s to working: %v", filename, err)
			continue
		}

		// Update status to working
		updateJobStatus(workingPath, "working", nil)

		// Start processing in a goroutine
		ctx, cancel := context.WithCancel(context.Background())
		jobID := filename[:len(filename)-4]
		activeContexts.Store(jobID, jobInfo{cancel: cancel, projectName: p.Name})
		go processJob(ctx, workingPath, p)

		// Increment job count for the project if it's not a docs job
		if !isDocsJob {
			count, _ := readJobCount(p)
			count++
			if count >= 10 {
				log.Printf("Project %s reached 10 jobs, enqueuing docs job", p.Name)
				writeJobCount(p, 0)
				_, err := EnqueueJob(p, JobRequest{
					Project: p.Name,
					Prompt:  "/bdoc-update",
				})
				if err != nil {
					log.Printf("Failed to enqueue automatic docs job for %s: %v", p.Name, err)
				}
			} else {
				writeJobCount(p, count)
			}
		}

		// Since schedule() only starts "build" jobs (via scripts/work),
		// we increment both global and build counters.
		totalWorkingGlobal++
		totalWorkingBuild++
		if totalWorkingGlobal >= MaxGlobalContainers || totalWorkingBuild >= MaxGlobalBuildContainers {
			break
		}
	}
}

func readJobCount(project Project) (int, error) {
	path := filepath.Join(project.Path(), "job_count")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}
	return count, nil
}

func writeJobCount(project Project, count int) error {
	path := filepath.Join(project.Path(), "job_count")
	return os.WriteFile(path, []byte(strconv.Itoa(count)), 0644)
}

func StopJob(jobID string) bool {
	if val, ok := activeContexts.Load(jobID); ok {
		info := val.(jobInfo)
		info.cancel()

		// Also stop the container explicitly
		containerName := fmt.Sprintf("overdrive-worker-%s-%s", info.projectName, jobID)
		go func() {
			// Try to stop the container gracefully, then kill if it takes too long
			exec.Command("podman", "stop", "-t", "2", containerName).Run()
		}()

		return true
	}
	return false
}

func CancelJob(projectName, jobID string) error {
	project, err := GetProject(projectName)
	if err != nil {
		return err
	}

	pendingPath := filepath.Join(project.JobsPath("pending"), jobID+".yml")
	cancelledPath := filepath.Join(project.JobsPath("cancelled"), jobID+".yml")

	// Ensure cancelled directory exists
	if err := os.MkdirAll(project.JobsPath("cancelled"), 0755); err != nil {
		return fmt.Errorf("failed to create cancelled directory: %w", err)
	}

	// Read job to update its status
	job, err := readJob(pendingPath)
	if err != nil {
		return fmt.Errorf("job not found in pending status: %w", err)
	}

	job.Status = "cancelled"
	now := time.Now()
	job.CompletedAt = &now
	job.Error = "Job cancelled by user"

	// Write updated job
	if err := writeJob(pendingPath, job); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Move to cancelled
	if err := os.Rename(pendingPath, cancelledPath); err != nil {
		return fmt.Errorf("failed to move job to cancelled: %w", err)
	}

	// Update Todo status if linked
	if err := UpdateTodoStatusFromJob(projectName, job); err != nil {
		log.Printf("Failed to update todo status for cancelled job %s: %v", jobID, err)
	}

	return nil
}

func processJob(ctx context.Context, filePath string, project Project) {
	filename := filepath.Base(filePath)
	jobID := filename[:len(filename)-4] // remove .yml

	defer activeContexts.Delete(jobID)

	// Read job details
	job, err := readJob(filePath)
	if err != nil {
		log.Printf("Failed to read job %s: %v", jobID, err)
		MoveToCrash(filePath, "Failed to read job file", project)
		return
	}

	// Setup logging
	logDir := project.LogsPath(jobID)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory %s: %v", logDir, err)
	}

	logFile, err := os.Create(filepath.Join(logDir, "worker.log"))
	if err != nil {
		log.Printf("Failed to create log file: %v", err)
	} else {
		defer logFile.Close()
	}

	// Run work script
	log.Printf("Job %s running scripts/work", jobID)
	if logFile != nil {
		fmt.Fprintf(logFile, "--- Running scripts/work ---\n")
	}

	// Use sh -c to allow complex commands
	cmd := exec.CommandContext(ctx, "./scripts/work")
	cmd.Env = os.Environ() // Inherit PATH etc

	// Environment variables
	envMap := map[string]string{
		"JOB_ID":           job.ID,
		"PROJECT_NAME":     project.Name,
		"REPO_URL":         job.Request.RepoURL,
		"SSH_KEY":          project.SSHKey,
		"SSH_CONFIG":       project.SSHConfig,
		"PRIMARY_BRANCH":   project.PrimaryBranch,
		"BRANCH_PARENT":    job.Request.BranchParent,
		"COMMIT_MSG":       job.Request.CommitMsg,
		"PROMPT":           job.Request.Prompt,
		"MODEL":            job.Request.Model,
		"LLM_PROVIDER":     project.Build.LLM.Provider,
		"HARNESS_PROVIDER": project.Build.Harness.Provider,
		"AUTH_JSON":        project.GenerateAuthJSON(project.Build),
		"DEPENDENCIES":     project.Dependencies,
		"CONTAINER_TYPE":   "build",
	}

	if timeout := os.Getenv("WORKER_TIMEOUT"); timeout != "" {
		envMap["TIMEOUT"] = timeout
	}

	// Build final env list
	for k, v := range envMap {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	// Project custom environment variables
	var envNames []string
	for k, v := range project.EnvVars {
		cmd.Env = append(cmd.Env, k+"="+v)
		envNames = append(envNames, k)
	}
	if len(envNames) > 0 {
		cmd.Env = append(cmd.Env, "WORKER_ENV_NAMES="+strings.Join(envNames, ","))
	}

	if logFile != nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	err = cmd.Run()
	exitCode := 0
	if err != nil {
		if ctx.Err() != nil {
			exitCode = 130 // Signal interruption
			log.Printf("Job %s was cancelled", jobID)
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
		log.Printf("Job %s failed with exit code %d: %v", jobID, exitCode, err)
	}

	if exitCode == 0 {
		log.Printf("Job %s completed successfully", jobID)
	}

	// Update job with result
	now := time.Now()
	job.CompletedAt = &now
	job.ExitCode = &exitCode

	// Read related commit if it was captured
	if commitData, err := os.ReadFile(filepath.Join(logDir, "related_commit")); err == nil {
		job.RelatedCommit = strings.TrimSpace(string(commitData))
	}

	// Read test results if they exist
	if testStatusData, err := os.ReadFile(filepath.Join(logDir, "test_status")); err == nil {
		job.TestStatus = strings.TrimSpace(string(testStatusData))
	}
	if testOutputData, err := os.ReadFile(filepath.Join(logDir, "test_output")); err == nil {
		job.TestOutput = string(testOutputData)
	}

	if exitCode == 2 {
		job.Status = "no-op"
	} else if ctx.Err() != nil {
		job.Status = "stopped"
		job.Error = "Job stopped by user"
	} else if exitCode == 124 {
		job.Status = "timeout"
		job.Error = "Job timed out"
	} else if exitCode != 0 {
		job.Status = "crash"
		job.Error = fmt.Sprintf("Process exited with code %d", exitCode)
		PauseProject(project.Name)
	} else {
		job.Status = "done"
	}

	// Write updated YAML
	if err := writeJob(filePath, job); err != nil {
		log.Printf("Failed to update job %s: %v", jobID, err)
	}

	// Update Todo status if linked
	if err := UpdateTodoStatusFromJob(project.Name, job); err != nil {
		log.Printf("Failed to update todo status for job %s: %v", jobID, err)
	}

	// Move file
	destDir := "done"
	if exitCode == 0 {
		destDir = "done"
	} else if exitCode == 2 {
		destDir = "no-op"
	} else if ctx.Err() != nil {
		destDir = "stopped"
	} else if exitCode == 124 {
		destDir = "timeout"
	} else if exitCode != 0 {
		destDir = "crash"
	}
	destPath := filepath.Join(project.JobsPath(destDir), filename)
	if err := os.Rename(filePath, destPath); err != nil {
		log.Printf("Failed to move job %s to %s: %v", jobID, destDir, err)
	}
}

func readJob(path string) (*Job, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var job Job
	if err := yaml.Unmarshal(data, &job); err != nil {
		return nil, err
	}

	// If job is working or pending, try to read sub_status from logs
	if (job.Status == "working" || job.Status == "pending") && job.Project != "" {
		subStatusPath := filepath.Join("projects", job.Project, "logs", job.ID, "sub_status")
		if data, err := os.ReadFile(subStatusPath); err == nil {
			job.SubStatus = strings.TrimSpace(string(data))
		}
	}

	// Always try to read test results if they exist
	if job.Project != "" && job.ID != "" {
		testStatusPath := filepath.Join("projects", job.Project, "logs", job.ID, "test_status")
		if data, err := os.ReadFile(testStatusPath); err == nil {
			job.TestStatus = strings.TrimSpace(string(data))
		}
		testOutputPath := filepath.Join("projects", job.Project, "logs", job.ID, "test_output")
		if data, err := os.ReadFile(testOutputPath); err == nil {
			job.TestOutput = string(data)
		}
	}

	return &job, nil
}

func writeJob(path string, job *Job) error {
	data, err := yaml.Marshal(job)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func updateJobStatus(path string, status string, errParams error) {
	job, err := readJob(path)
	if err != nil {
		log.Printf("Failed to read job for status update: %v", err)
		return
	}
	job.Status = status
	if status == "working" {
		now := time.Now()
		job.StartedAt = &now
	}
	if errParams != nil {
		job.Error = errParams.Error()
	}
	writeJob(path, job)
}

func MoveToCrash(path string, reason string, project Project) {
	job, _ := readJob(path)
	if job != nil {
		job.Status = "crash"
		job.Error = reason
		now := time.Now()
		job.CompletedAt = &now
		writeJob(path, job)
	}

	PauseProject(project.Name)

	filename := filepath.Base(path)
	os.Rename(path, filepath.Join(project.JobsPath("crash"), filename))
}
