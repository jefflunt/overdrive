package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type SettingsPageData struct {
	CurrentPath              string
	Project                  Project
	MaxGlobalContainers      int
	MaxGlobalBuildContainers int
	MaxGlobalChatContainers  int
	MaxGlobalCmdContainers   int
}

func HandleSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Ensure global variables are up to date
	if _, err := LoadGlobalSettings(); err != nil {
		log.Printf("Error loading global settings: %v", err)
	}

	data := SettingsPageData{
		CurrentPath:              r.URL.Path,
		MaxGlobalContainers:      MaxGlobalContainers,
		MaxGlobalBuildContainers: MaxGlobalBuildContainers,
		MaxGlobalChatContainers:  MaxGlobalChatContainers,
		MaxGlobalCmdContainers:   MaxGlobalCmdContainers,
	}

	projectName := getProjectName(r)
	if projectName != "" {
		if project, err := GetProject(projectName); err == nil && project != nil {
			data.Project = *project
		}
	}

	tmpl, err := parseTemplate("api/templates/settings.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func HandleSaveGlobalSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var settings GlobalSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := SaveGlobalSettings(&settings); err != nil {
		log.Printf("Error saving global settings: %v", err)
		http.Error(w, "Failed to save global settings: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Global settings saved successfully: %+v", settings)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func HandleRebuildRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Rebuild and restart api triggered...")

	// Run the script in a separate goroutine so we can return a response
	go func() {
		time.Sleep(500 * time.Millisecond) // Give time for the response to go out
		cmd := exec.Command("./scripts/rebuild-and-restart")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to run rebuild-and-restart: %v", err)
		}
	}()

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "Rebuild and restart api initiated. The server will be unavailable for a few moments.")
}

func HandleRebuildRestartScheduler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Rebuild and restart scheduler triggered...")

	// Run the script in a separate goroutine so we can return a response
	go func() {
		// Asynchronously stop any running containers
		stopCmd := exec.Command("podman", "stop", "--all")
		if err := stopCmd.Start(); err != nil {
			log.Printf("Failed to initiate podman stop: %v", err)
		}

		time.Sleep(500 * time.Millisecond) // Give time for the response to go out
		cmd := exec.Command("./scripts/rebuild-and-restart-scheduler")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to run rebuild-and-restart-scheduler: %v", err)
		}
	}()

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "Rebuild and restart scheduler initiated.")
}

func HandleHealthInfo(w http.ResponseWriter, r *http.Request) {
	containers, err := GetRunningContainersInfo()
	if err != nil {
		log.Printf("Error getting container info: %v", err)
	}

	apiUptime := time.Since(ApiStartTime)
	schedulerUptime := time.Duration(0)
	if st, err := GetSchedulerStartTime(); err == nil {
		schedulerUptime = time.Since(st)
	}

	data := map[string]interface{}{
		"Containers":      containers,
		"ApiUptime":       FormatUptime(apiUptime),
		"SchedulerUptime": FormatUptime(schedulerUptime),
	}

	if r.Header.Get("HX-Request") == "true" {
		// Return HTML snippet for HTMX
		fmt.Fprintf(w, `
			<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
				<div class="bg-slate-50 dark:bg-slate-800/50 p-4 rounded border border-slate-200 dark:border-slate-700">
					<div class="text-xs text-slate-500 uppercase font-bold mb-1">API Uptime</div>
					<div class="text-lg font-mono text-slate-900 dark:text-slate-200">%s</div>
				</div>
				<div class="bg-slate-50 dark:bg-slate-800/50 p-4 rounded border border-slate-200 dark:border-slate-700">
					<div class="text-xs text-slate-500 uppercase font-bold mb-1">Scheduler Uptime</div>
					<div class="text-lg font-mono text-slate-900 dark:text-slate-200">%s</div>
				</div>
				<div class="bg-slate-50 dark:bg-slate-800/50 p-4 rounded border border-slate-200 dark:border-slate-700">
					<div class="text-xs text-slate-500 uppercase font-bold mb-1">Active Containers</div>
					<div class="text-lg font-mono text-slate-900 dark:text-slate-200">%d</div>
				</div>
			</div>
		`, data["ApiUptime"], data["SchedulerUptime"], len(containers))

		if len(containers) > 0 {
			fmt.Fprint(w, `
				<div class="overflow-x-auto border border-slate-200 dark:border-slate-700 rounded">
					<table class="min-w-full divide-y divide-slate-200 dark:divide-slate-700">
						<thead class="bg-slate-50 dark:bg-slate-800/50">
							<tr>
								<th class="px-4 py-2 text-left text-xs font-bold text-slate-500 uppercase whitespace-nowrap">Name</th>
								<th class="px-4 py-2 text-left text-xs font-bold text-slate-500 uppercase whitespace-nowrap">Project</th>
								<th class="px-4 py-2 text-left text-xs font-bold text-slate-500 uppercase whitespace-nowrap">Type</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-200 dark:divide-slate-700 bg-white dark:bg-slate-900/50">
			`)
			for _, c := range containers {
				project := c.Project
				if project == "" {
					project = "-"
				}
				cType := c.Type
				if cType == "" {
					cType = "-"
				}
				fmt.Fprintf(w, `
							<tr>
								<td class="px-4 py-2 text-sm font-mono text-slate-700 dark:text-slate-300 whitespace-nowrap">%s</td>
								<td class="px-4 py-2 text-sm text-slate-600 dark:text-slate-400 whitespace-nowrap">%s</td>
								<td class="px-4 py-2 text-sm text-slate-600 dark:text-slate-400 whitespace-nowrap">%s</td>
							</tr>
				`, c.Name, project, cType)
			}
			fmt.Fprint(w, `
						</tbody>
					</table>
				</div>
			`)
		} else {
			fmt.Fprint(w, `<div class="text-sm text-slate-500 italic">No active containers found.</div>`)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
