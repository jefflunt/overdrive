package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"overdrive/api"
)

func main() {
	api.ApiStartTime = time.Now()
	// Ensure projects directory exists
	if err := os.MkdirAll("projects", 0755); err != nil {
		log.Fatalf("Failed to create projects directory: %v", err)
	}

	// Load global settings
	if _, err := api.LoadGlobalSettings(); err != nil {
		log.Printf("Warning: Failed to load global settings: %v", err)
	}

	// Create .keep file in projects directory
	os.WriteFile("projects/.keep", []byte(""), 0644)

	mux := http.NewServeMux()

	// Static routes
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/icons/icon-192.png")
	})

	// Healthcheck endpoint
	mux.HandleFunc("/up", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.HandleSaveProject(w, r)
		} else {
			api.HandleListProjects(w, r)
		}
	})

	mux.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Check for nested jobs routes
		// Pattern: /projects/{project}/jobs...
		parts := strings.Split(strings.TrimPrefix(path, "/projects/"), "/")
		if len(parts) >= 2 {
			if parts[1] == "jobs" {
				if len(parts) == 2 || parts[2] == "" {
					if r.Method == http.MethodPost {
						api.HandleSubmitJob(w, r)
					} else {
						api.HandleListJobs(w, r)
					}
					return
				}

				// Sub-routes: /projects/{project}/jobs/{sub}/{id}
				sub := parts[2]
				switch sub {
				case "logs":
					api.HandleViewLogs(w, r)
				case "logs-partial":
					api.HandlePartialLogs(w, r)
				case "tail":
					api.HandleTailLogs(w, r)
				case "updates":
					api.HandleJobUpdates(w, r)
				case "stop":
					api.HandleStopJob(w, r)
				case "cancel":
					api.HandleCancelJob(w, r)
				case "revert":
					api.HandleRevertJob(w, r)
				case "diff":
					api.HandleJobDiff(w, r)
				case "delete":
					api.HandleDeleteJob(w, r)
				default:
					http.NotFound(w, r)
				}
				return
			}
			if parts[1] == "cmds" {
				if r.Method == http.MethodPost {
					api.HandleRunProjectCmd(w, r)
					return
				}
			}
			if parts[1] == "config" {
				if r.Method == http.MethodGet {
					api.HandleGetProjectConfig(w, r)
					return
				}
			}
			if parts[1] == "chat" {
				api.HandleProjectChat(w, r)
				return
			}
			if parts[1] == "resume" {
				if r.Method == http.MethodPost {
					api.HandleResumeProject(w, r)
					return
				}
			}
			if parts[1] == "todos" {
				if len(parts) == 2 || parts[2] == "" {
					if r.Method == http.MethodPost {
						api.HandleCreateTodo(w, r)
					} else {
						api.HandleListTodos(w, r)
					}
					return
				}
				// /projects/{project}/todos/{id}
				if len(parts) > 3 {
					// /projects/{project}/todos/{id}/submit
					if parts[3] == "submit" {
						api.HandleSubmitTodo(w, r)
						return
					}
					// /projects/{project}/todos/{id}/star
					if parts[3] == "star" {
						api.HandleStarTodo(w, r)
						return
					}
				}
				if r.Method == http.MethodPut {
					api.HandleUpdateTodo(w, r)
				} else if r.Method == http.MethodDelete {
					api.HandleDeleteTodo(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
				return
			}
		}

		if r.Method == http.MethodDelete {
			api.HandleDeleteProject(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Keep old routes but they are now secondary
	mux.HandleFunc("/jobs/logs/", api.HandleViewLogs)
	mux.HandleFunc("/jobs/logs-partial/", api.HandlePartialLogs)
	mux.HandleFunc("/jobs/tail/", api.HandleTailLogs)
	mux.HandleFunc("/jobs/updates/", api.HandleJobUpdates)
	mux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.HandleSubmitJob(w, r)
		} else {
			api.HandleListJobs(w, r)
		}
	})

	// Help routes
	mux.HandleFunc("/help", api.HandleHelp)

	// Dashboard routes
	mux.HandleFunc("/dashboard", api.HandleDashboard)
	mux.HandleFunc("/api/dashboard/todos", api.HandleGetDashboardTodos)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		api.HandleListProjects(w, r)
	})

	// Settings route
	mux.HandleFunc("/settings", api.HandleSettings)
	mux.HandleFunc("/settings/global/save", api.HandleSaveGlobalSettings)
	mux.HandleFunc("/health-info", api.HandleHealthInfo)
	mux.HandleFunc("/rebuild-restart", api.HandleRebuildRestart)
	mux.HandleFunc("/rebuild-restart-scheduler", api.HandleRebuildRestartScheduler)

	port := "3281"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
