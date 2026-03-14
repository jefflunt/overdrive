package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// JobListPage data for template
type JobListPage struct {
	CurrentPath      string
	Project          Project
	Jobs             []Job
	Offset           int
	Limit            int
	NextOffset       int
	HasMore          bool
	TotalJobs        int
	AllStatuses      []string
	SelectedStatuses []string
	SearchQuery      string
	IsHX             bool
	InitialPrompt    string
}

type ProjectListPage struct {
	CurrentPath string
	Project     Project
	Projects    []Project
}

type HelpPage struct {
	CurrentPath string
	CurrentPage string
	Doc         *DocFile
	Project     Project
}

type DocFile struct {
	Name    string
	Path    string
	Title   string
	Content string
}

func parseTemplate(filename string) (*template.Template, error) {
	jsonMarshal := func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		return string(b), err
	}

	funcMap := template.FuncMap{
		"json":           jsonMarshal,
		"formatPrompt":   FormatPrompt,
		"markdown":       MarkdownToHtml,
		"renderMarkdown": MarkdownToHtml,
		"listProjects":   ListProjects,
		"hasPrefix":      strings.HasPrefix,
		"hasSuffix":      strings.HasSuffix,
		"contains":       strings.Contains,
		"join":           strings.Join,
		"add": func(a, b int) int {
			return a + b
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"getBuildNumber": func() string {
			out, err := exec.Command("sh", "-c", "git log --oneline | wc -l").Output()
			if err != nil {
				return "unknown"
			}
			return "b" + strings.TrimSpace(string(out))
		},
	}

	tmpl := template.New("layout.html").Funcs(funcMap)

	if filename != "" {
		return tmpl.ParseFiles("api/templates/layout.html", filename)
	}
	return tmpl.ParseFiles("api/templates/layout.html")
}

func getProjectName(r *http.Request) string {
	if name := r.URL.Query().Get("project"); name != "" {
		return name
	}
	parts := strings.Split(r.URL.Path, "/")
	for i, part := range parts {
		if part == "projects" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

type LogPageData struct {
	CurrentPath string
	Project     Project
	JobID       string
	Logs        template.HTML
	Job         *Job
}

type ViewLogsData struct {
	CurrentPath  string
	Project      Project
	JobID        string
	ExitCodeHtml template.HTML
	Logs         template.HTML
	Job          *Job
}

type DiffPageData struct {
	CurrentPath string
	Project     Project
	JobID       string
	Diff        template.HTML
	Job         *Job
}
