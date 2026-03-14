package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func HandleListTodos(w http.ResponseWriter, r *http.Request) {
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

	// Serve HTML if requested (browser navigation)
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		tmpl, err := parseTemplate("api/templates/todos.html")
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Project": *project,
		}

		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	var todos []Todo
	if project.TodoProvider == "jira" {
		todos, err = FetchJiraIssues(project.Jira)
	} else if project.TodoProvider == "github" {
		todos, err = FetchGitHubIssues(project.GitHub)
	} else {
		todos, err = LoadTodos(projectName)
	}

	if err != nil {
		http.Error(w, "Failed to load todos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func HandleCreateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	if project.TodoProvider != "native" {
		http.Error(w, "Creation is disabled in "+project.TodoProvider+" mode", http.StatusForbidden)
		return
	}

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	todo.ID = GenerateJobID() // Reuse existing ID generator
	todo.ProjectID = projectName
	todo.Status = "draft"
	todo.CreatedAt = time.Now().Unix()

	todos, err := LoadTodos(projectName)
	if err != nil {
		http.Error(w, "Failed to load todos", http.StatusInternalServerError)
		return
	}

	updatedTodos, err := AddTodo(todos, todo)
	if err != nil {
		http.Error(w, "Failed to add todo: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := SaveTodos(projectName, updatedTodos); err != nil {
		http.Error(w, "Failed to save todos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func HandleUpdateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	if project.TodoProvider != "native" {
		http.Error(w, "Update is disabled in "+project.TodoProvider+" mode", http.StatusForbidden)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	// /projects/{project}/todos/{id}
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	todoID := pathParts[4]

	var update Todo
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	update.ID = todoID // Ensure ID matches URL

	todos, err := LoadTodos(projectName)
	if err != nil {
		http.Error(w, "Failed to load todos", http.StatusInternalServerError)
		return
	}

	existing := FindTodo(todos, todoID)
	if existing == nil {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	if existing.Status == "submitted" {
		http.Error(w, "Cannot update submitted todo", http.StatusForbidden)
		return
	}

	newTodos, found := UpdateTodoInTree(todos, update)
	if !found {
		http.Error(w, "Todo not found for update", http.StatusNotFound)
		return
	}

	if err := SaveTodos(projectName, newTodos); err != nil {
		http.Error(w, "Failed to save todos", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HandleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	if project.TodoProvider != "native" {
		http.Error(w, "Deletion is disabled in "+project.TodoProvider+" mode", http.StatusForbidden)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	todoID := pathParts[4]

	todos, err := LoadTodos(projectName)
	if err != nil {
		http.Error(w, "Failed to load todos", http.StatusInternalServerError)
		return
	}

	newTodos := DeleteTodoFromTree(todos, todoID)

	if err := SaveTodos(projectName, newTodos); err != nil {
		http.Error(w, "Failed to save todos", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HandleStarTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	if project.TodoProvider != "native" {
		http.Error(w, "Starring is disabled in "+project.TodoProvider+" mode", http.StatusForbidden)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	// /projects/{project}/todos/{id}/star
	if len(pathParts) < 6 || pathParts[5] != "star" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	todoID := pathParts[4]

	var update struct {
		Starred bool `json:"starred"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	todos, err := LoadTodos(projectName)
	if err != nil {
		http.Error(w, "Failed to load todos", http.StatusInternalServerError)
		return
	}

	existing := FindTodo(todos, todoID)
	if existing == nil {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	newTodos, found := UpdateTodoStar(todos, todoID, update.Starred)
	if !found {
		http.Error(w, "Todo not found for update", http.StatusNotFound)
		return
	}

	if err := SaveTodos(projectName, newTodos); err != nil {
		http.Error(w, "Failed to save todos", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HandleSubmitTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	pathParts := strings.Split(r.URL.Path, "/")
	// /projects/{project}/todos/{id}/submit
	if len(pathParts) < 6 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	todoID := pathParts[4]

	var todos []Todo
	if project.TodoProvider == "jira" {
		todos, err = FetchJiraIssues(project.Jira)
	} else if project.TodoProvider == "github" {
		todos, err = FetchGitHubIssues(project.GitHub)
	} else {
		todos, err = LoadTodos(projectName)
	}

	if err != nil {
		http.Error(w, "Failed to load todos", http.StatusInternalServerError)
		return
	}

	todo := FindTodo(todos, todoID)
	if todo == nil {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	// Validate: only leaf nodes
	if len(todo.Children) > 0 {
		http.Error(w, "Cannot submit parent node", http.StatusBadRequest)
		return
	}

	// Update status to "In Progress" if enabled
	if project.TodoProvider == "jira" {
		statusPickup := project.Jira.StatusPickup
		if statusPickup == "" {
			statusPickup = "In Progress"
		}
		err := UpdateJiraIssueStatus(project.Jira, todo.ID, statusPickup)
		if err != nil {
			http.Error(w, "Failed to update Jira status: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if project.TodoProvider == "github" {
		statusPickup := project.GitHub.StatusPickup
		if statusPickup != "" {
			err := UpdateGitHubIssueStatus(project.GitHub, todo.ID, statusPickup)
			if err != nil {
				http.Error(w, "Failed to update GitHub status: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		// Validate: status == draft
		if todo.Status != "draft" {
			http.Error(w, "Only draft todos can be submitted", http.StatusBadRequest)
			return
		}
	}

	// Create Job
	req := JobRequest{
		Project:      projectName,
		RepoURL:      project.RepoURL,
		BranchParent: project.PrimaryBranch,
		Prompt:       fmt.Sprintf("/bdoc-engineer # title\n\n%s\n\n## details\n\n%s", todo.Title, todo.Description), // Combine title and description
		Model:        project.Build.LLM.Model,
		CommitMsg:    todo.Title,
		ReferenceID:  todoID,
	}

	jobID, err := EnqueueJob(*project, req)
	if err != nil {
		http.Error(w, "Failed to enqueue job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update Todo Status (if not Jira)
	if project.TodoProvider != "jira" {
		todos, _ = UpdateTodoStatus(todos, todoID, "submitted", jobID)
		if err := SaveTodos(projectName, todos); err != nil {
			http.Error(w, "Failed to save todo status", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"jobId": jobID})
}
