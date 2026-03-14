package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"sort"
	"time"
)

type DashboardData struct {
	Projects    map[string][]Todo `json:"projects"`
	RunningJobs []Job             `json:"running_jobs"`
	RecentJobs  []Job             `json:"recent_jobs"`
}

func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := parseTemplate("api/templates/dashboard.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	projects, _ := ListProjects()

	data := struct {
		CurrentPath string
		Projects    []Project
		Project     Project
	}{
		CurrentPath: "/dashboard",
		Projects:    projects,
		Project:     Project{},
	}

	tmpl.ExecuteTemplate(w, "layout.html", data)
}

func HandleGetDashboardTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projects, err := ListProjects()
	if err != nil {
		http.Error(w, "Failed to list projects", http.StatusInternalServerError)
		return
	}

	dashboardData := DashboardData{
		Projects:    make(map[string][]Todo),
		RunningJobs: []Job{},
		RecentJobs:  []Job{},
	}

	cutoff := time.Now().Add(-24 * time.Hour)

	for _, p := range projects {
		if p.TodoProvider == "jira" {
			todos, err := FetchJiraIssues(p.Jira)
			if err == nil {
				dashboardData.Projects[p.Name] = todos
			}
		} else if p.TodoProvider == "github" {
			todos, err := FetchGitHubIssues(p.GitHub)
			if err == nil {
				dashboardData.Projects[p.Name] = todos
			}
		} else {
			todos, err := LoadTodos(p.Name)
			if err == nil {
				dashboardData.Projects[p.Name] = todos
			}
		}

		// Collect running and pending jobs
		for _, status := range []string{"working", "pending"} {
			files, _ := filepath.Glob(filepath.Join(p.JobsPath(status), "*.yml"))
			for _, f := range files {
				if job, err := readJob(f); err == nil && job != nil {
					dashboardData.RunningJobs = append(dashboardData.RunningJobs, *job)
				}
			}
		}

		// Collect recent finished jobs
		for _, status := range []string{"done", "crash", "no-op", "timeout", "stopped", "undone", "cancelled"} {
			files, _ := filepath.Glob(filepath.Join(p.JobsPath(status), "*.yml"))
			for _, f := range files {
				if job, err := readJob(f); err == nil && job != nil {
					// Check if it's within last 24 hours (using CreatedAt as per request: "most recently created jobs first")
					// Actually, "recent jobs that finished running" might mean we should check CompletedAt,
					// but "most recently created jobs first" implies CreatedAt for sorting.
					// I'll check if it finished in last 24h OR was created in last 24h.
					// The prompt says: "this list should show all jobs run within the last 24 hours"
					// and "sorted in reverse chronological order (most recently created jobs first)".
					if job.CreatedAt.After(cutoff) || (job.CompletedAt != nil && job.CompletedAt.After(cutoff)) {
						dashboardData.RecentJobs = append(dashboardData.RecentJobs, *job)
					}
				}
			}
		}
	}

	// Sort running jobs by created at descending
	sort.Slice(dashboardData.RunningJobs, func(i, j int) bool {
		return dashboardData.RunningJobs[i].CreatedAt.After(dashboardData.RunningJobs[j].CreatedAt)
	})

	// Sort recent jobs by created at descending
	sort.Slice(dashboardData.RecentJobs, func(i, j int) bool {
		return dashboardData.RecentJobs[i].CreatedAt.After(dashboardData.RecentJobs[j].CreatedAt)
	})

	// Limit to 7
	if len(dashboardData.RecentJobs) > 7 {
		dashboardData.RecentJobs = dashboardData.RecentJobs[:7]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboardData)
}
