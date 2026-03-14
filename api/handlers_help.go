package api

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func HandleHelp(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	if from == "" {
		from = r.URL.Query().Get("page")
	}

	docName := ""
	if from != "" {
		// Map path to doc
		if strings.Contains(from, "/jobs") {
			docName = "JOBS.md"
		} else if strings.Contains(from, "/chat") {
			docName = "CHATS.md"
		} else if strings.Contains(from, "/settings") {
			docName = "CONFIG.md"
		} else if strings.Contains(from, "/todos") {
			docName = "TODOS.md"
		}
	}

	// Explicit doc request via ?doc=...
	if d := r.URL.Query().Get("doc"); d != "" {
		docName = d
	}

	// Load selected doc
	var selectedDoc *DocFile
	if docName != "" {
		path := filepath.Join("help_docs", docName)
		if content, err := os.ReadFile(path); err == nil {
			title := strings.TrimSuffix(docName, ".md")
			lines := strings.Split(string(content), "\n")
			if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
				title = strings.TrimPrefix(lines[0], "# ")
			}
			selectedDoc = &DocFile{
				Name:    docName,
				Path:    path,
				Title:   title,
				Content: string(content),
			}
		}
	}

	data := HelpPage{
		CurrentPath: r.URL.Path,
		CurrentPage: from,
		Doc:         selectedDoc,
	}

	tmpl, err := parseTemplate("api/templates/help.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}
