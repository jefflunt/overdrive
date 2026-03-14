package api

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Todo struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	ParentID    string `json:"parentId,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"` // "draft", "submitted", "completed", "crashed"
	Starred     bool   `json:"starred,omitempty"`
	JobID       string `json:"jobId,omitempty"`
	Children    []Todo `json:"children,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
}

var todoMutex sync.Mutex

func GetTodosPath(project string) string {
	return filepath.Join("projects", project, "todos.json")
}

func LoadTodos(project string) ([]Todo, error) {
	todoMutex.Lock()
	defer todoMutex.Unlock()

	path := GetTodosPath(project)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Todo{}, nil
		}
		return nil, err
	}

	var todos []Todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, err
	}
	if todos == nil {
		return []Todo{}, nil
	}
	return todos, nil
}

func SaveTodos(project string, todos []Todo) error {
	todoMutex.Lock()
	defer todoMutex.Unlock()

	if todos == nil {
		todos = []Todo{}
	}

	path := GetTodosPath(project)
	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func FindTodo(todos []Todo, id string) *Todo {
	for i := range todos {
		if todos[i].ID == id {
			return &todos[i]
		}
		if found := FindTodo(todos[i].Children, id); found != nil {
			return found
		}
	}
	return nil
}

func FindTodoByJobID(todos []Todo, jobID string) *Todo {
	for i := range todos {
		if todos[i].JobID == jobID {
			return &todos[i]
		}
		if found := FindTodoByJobID(todos[i].Children, jobID); found != nil {
			return found
		}
	}
	return nil
}

func AddTodo(todos []Todo, newTodo Todo) ([]Todo, error) {
	if newTodo.ParentID == "" {
		return append(todos, newTodo), nil
	}

	// Helper function to find parent and append child
	var addToParent func([]Todo) ([]Todo, bool)
	addToParent = func(list []Todo) ([]Todo, bool) {
		for i := range list {
			if list[i].ID == newTodo.ParentID {
				list[i].Children = append(list[i].Children, newTodo)
				return list, true
			}
			if len(list[i].Children) > 0 {
				if updatedChildren, found := addToParent(list[i].Children); found {
					list[i].Children = updatedChildren
					return list, true
				}
			}
		}
		return list, false
	}

	updatedTodos, found := addToParent(todos)
	if !found {
		return todos, errors.New("parent not found")
	}
	return updatedTodos, nil
}

func UpdateTodoInTree(todos []Todo, updated Todo) ([]Todo, bool) {
	for i := range todos {
		if todos[i].ID == updated.ID {
			todos[i].Title = updated.Title
			todos[i].Description = updated.Description
			return todos, true
		}
		if len(todos[i].Children) > 0 {
			if newChildren, found := UpdateTodoInTree(todos[i].Children, updated); found {
				todos[i].Children = newChildren
				return todos, true
			}
		}
	}
	return todos, false
}

func UpdateTodoStatus(todos []Todo, id string, status string, jobID string) ([]Todo, bool) {
	for i := range todos {
		if todos[i].ID == id {
			todos[i].Status = status
			if jobID != "" {
				todos[i].JobID = jobID
			}
			return todos, true
		}
		if len(todos[i].Children) > 0 {
			if newChildren, found := UpdateTodoStatus(todos[i].Children, id, status, jobID); found {
				todos[i].Children = newChildren
				return todos, true
			}
		}
	}
	return todos, false
}

func UpdateTodoStar(todos []Todo, id string, starred bool) ([]Todo, bool) {
	for i := range todos {
		if todos[i].ID == id {
			todos[i].Starred = starred
			return todos, true
		}
		if len(todos[i].Children) > 0 {
			if newChildren, found := UpdateTodoStar(todos[i].Children, id, starred); found {
				todos[i].Children = newChildren
				return todos, true
			}
		}
	}
	return todos, false
}

func DeleteTodoFromTree(todos []Todo, id string) []Todo {
	var newTodos []Todo
	for _, t := range todos {
		if t.ID == id {
			continue // Skip this one (delete)
		}
		if len(t.Children) > 0 {
			t.Children = DeleteTodoFromTree(t.Children, id)
		}
		newTodos = append(newTodos, t)
	}
	return newTodos
}

func UpdateTodoStatusFromJob(projectName string, job *Job) error {
	project, err := GetProject(projectName)
	if err != nil {
		return err
	}

	var todos []Todo
	if project.TodoProvider == "jira" {
		todos, err = FetchJiraIssues(project.Jira)
		if err != nil {
			return err
		}
	} else {
		todos, err = LoadTodos(projectName)
		if err != nil {
			return err
		}
	}

	todoStatus := ""
	jiraStatus := ""
	switch job.Status {
	case "done":
		todoStatus = "completed"
		jiraStatus = project.Jira.StatusDone
		if jiraStatus == "" {
			jiraStatus = "Done"
		}
	case "crash", "timeout", "stopped", "undone", "cancelled":
		todoStatus = "crashed"
	default:
		return nil // No update needed for other statuses
	}

	// Find the todo
	var target *Todo
	if project.TodoProvider == "jira" && job.ReferenceID != "" {
		target = FindTodo(todos, job.ReferenceID)
	} else {
		target = FindTodoByJobID(todos, job.ID)
	}

	if target == nil {
		return nil // No todo linked to this job
	}

	if project.TodoProvider == "jira" && jiraStatus != "" {
		err := UpdateJiraIssueStatus(project.Jira, target.ID, jiraStatus)
		if err != nil {
			return err
		}
	}

	if project.TodoProvider != "jira" {
		// Update status locally
		todos, _ = UpdateTodoStatus(todos, target.ID, todoStatus, "")
		return SaveTodos(projectName, todos)
	}

	return nil
}
