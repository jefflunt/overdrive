package api

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type ChatPageData struct {
	CurrentPath  string
	Project      Project
	Chats        []Chat
	DeletedChats []Chat
	Current      *Chat
	IsGenerating bool
	ShowTrash    bool
}

func HandleProjectChat(w http.ResponseWriter, r *http.Request) {
	projectName := getProjectName(r)
	if projectName == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	project, err := GetProject(projectName)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	// /projects/{name}/chat/list
	// /projects/{name}/chat/messages/{id}
	// /projects/{name}/chat/send
	// /projects/{name}/chat/create
	// /projects/{name}/chat/delete/{id}

	if len(pathParts) > 4 {
		sub := pathParts[4]
		switch sub {
		case "list":
			HandleChatList(w, r, *project)
			return
		case "messages":
			if len(pathParts) > 5 {
				HandleChatMessages(w, r, *project, pathParts[5], nil)
				return
			}
		case "send":
			HandleChatSend(w, r, *project)
			return
		case "build":
			HandleChatBuild(w, r, *project)
			return
		case "create":
			HandleChatCreate(w, r, *project)
			return
		case "delete":
			if len(pathParts) > 5 {
				HandleChatDelete(w, r, *project, pathParts[5])
				return
			}
		case "restore":
			if len(pathParts) > 5 {
				HandleChatRestore(w, r, *project, pathParts[5])
				return
			}
		case "rename":
			if len(pathParts) > 5 {
				HandleChatRename(w, r, *project, pathParts[5])
				return
			}
		case "proxy":
			if len(pathParts) > 5 {
				HandleChatProxy(w, r, *project, pathParts[5])
				return
			}
		case "sync":
			if len(pathParts) > 5 {
				HandleChatSync(w, r, *project, pathParts[5])
				return
			}
		case "warm":
			if len(pathParts) > 5 {
				WarmChatSession(*project, pathParts[5])
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	}

	chats, _ := ListChats(projectName)
	showTrash := r.URL.Query().Get("trash") == "true"
	if showTrash {
		CleanupOldDeletedChats(projectName)
	}
	deletedChats, _ := ListDeletedChats(projectName)

	var currentChat *Chat
	chatID := r.URL.Query().Get("id")
	isGenerating := false

	if chatID != "" {
		log.Printf("[DEBUG] HandleProjectChat: loading chatID=%s", chatID)
		currentChat, _ = GetChat(projectName, chatID)
		WarmChatSession(*project, chatID)

		isGenerating = false
		session := GetChatSession(projectName, chatID)
		if session != nil && session.OpenCodeSessionID != "" {
			client := NewOpenCodeClient(session.Port)
			if generating, err := client.IsGenerating(session.OpenCodeSessionID); err == nil {
				isGenerating = generating
			}
		}

		// If we don't have a session yet but the last message is from the user,
		// we should still treat it as generating so the UI starts polling.
		if !isGenerating && currentChat != nil && len(currentChat.Messages) > 0 {
			if currentChat.Messages[len(currentChat.Messages)-1].Role == "user" {
				isGenerating = true
			}
		}

		if isGenerating && currentChat != nil && len(currentChat.Messages) > 0 {
			if currentChat.Messages[len(currentChat.Messages)-1].Role == "assistant" {
				currentChat.Messages = currentChat.Messages[:len(currentChat.Messages)-1]
			}
		}
	}

	data := ChatPageData{
		CurrentPath:  r.URL.Path,
		Project:      *project,
		Chats:        chats,
		DeletedChats: deletedChats,
		Current:      currentChat,
		IsGenerating: isGenerating,
		ShowTrash:    showTrash,
	}

	tmpl, err := parseTemplate("api/templates/chat.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func HandleChatList(w http.ResponseWriter, r *http.Request, project Project) {
	chats, _ := ListChats(project.Name)
	showTrash := r.URL.Query().Get("trash") == "true"
	if showTrash {
		CleanupOldDeletedChats(project.Name)
	}
	deletedChats, _ := ListDeletedChats(project.Name)

	var currentChat *Chat
	currentID := r.URL.Query().Get("id")
	if currentID == "" {
		currentURL := r.Header.Get("HX-Current-Url")
		if currentURL != "" {
			if u, err := url.Parse(currentURL); err == nil {
				currentID = u.Query().Get("id")
			}
		}
	}
	if currentID != "" {
		currentChat, _ = GetChat(project.Name, currentID)
	}

	tmpl, _ := parseTemplate("api/templates/chat.html")
	tmpl.ExecuteTemplate(w, "chat-sidebar-items", map[string]interface{}{
		"Chats":        chats,
		"DeletedChats": deletedChats,
		"Project":      project,
		"Current":      currentChat,
		"ShowTrash":    showTrash,
	})
}

func HandleChatMessages(w http.ResponseWriter, r *http.Request, project Project, chatID string, chat *Chat) {
	WarmChatSession(project, chatID)
	if chat == nil {
		var err error
		chat, err = GetChat(project.Name, chatID)
		if err != nil {
			log.Printf("[ERROR] HandleChatMessages: Chat %s not found for project %s: %v", chatID, project.Name, err)
			http.Error(w, "Chat not found", http.StatusNotFound)
			return
		}
	}

	isGenerating := false
	session := GetChatSession(project.Name, chatID)
	if session != nil && session.OpenCodeSessionID != "" {
		client := NewOpenCodeClient(session.Port)
		if generating, err := client.IsGenerating(session.OpenCodeSessionID); err == nil {
			isGenerating = generating
		}
	}

	// Fallback to treat as generating if last message is from user
	if !isGenerating && chat != nil && len(chat.Messages) > 0 {
		if chat.Messages[len(chat.Messages)-1].Role == "user" {
			isGenerating = true
		}
	}

	if isGenerating && chat != nil && len(chat.Messages) > 0 {
		if chat.Messages[len(chat.Messages)-1].Role == "assistant" {
			chat.Messages = chat.Messages[:len(chat.Messages)-1]
		}
	}

	tmpl, err := parseTemplate("api/templates/chat.html")
	if err != nil {
		log.Printf("[ERROR] HandleChatMessages: Failed to parse template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Current":      chat,
		"Project":      project,
		"IsGenerating": isGenerating,
	}

	w.Header().Set("Content-Type", "text/html")
	if r.URL.Query().Get("only_inner") == "true" {
		if err := tmpl.ExecuteTemplate(w, "chat-messages-inner", data); err != nil {
			log.Printf("[ERROR] HandleChatMessages: Failed to execute template: %v", err)
		}
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "chatListChanged")
	}
	if err := tmpl.ExecuteTemplate(w, "chat-messages", data); err != nil {
		log.Printf("[ERROR] HandleChatMessages: Failed to execute template: %v", err)
	}
}

func HandleChatCreate(w http.ResponseWriter, r *http.Request, project Project) {
	chatID := GenerateChatID()
	chat := Chat{
		ID:        chatID,
		Title:     "New Chat",
		Project:   project.Name,
		Messages:  []ChatMessage{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := SaveChat(&chat); err != nil {
		log.Printf("[ERROR] HandleChatCreate: Failed to save chat: %v", err)
		http.Error(w, "Failed to save chat", http.StatusInternalServerError)
		return
	}
	WarmChatSession(project, chatID)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Push-Url", fmt.Sprintf("/projects/%s/chat?id=%s", project.Name, chatID))
		HandleChatMessages(w, r, project, chatID, &chat)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/projects/%s/chat?id=%s", project.Name, chatID), http.StatusSeeOther)
}

func HandleChatProxy(w http.ResponseWriter, r *http.Request, project Project, chatID string) {
	session, err := GetOrCreateChatSession(project, chatID)
	if err != nil {
		http.Error(w, "Failed to start chat session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	targetURL := fmt.Sprintf("http://127.0.0.1:%s", session.Port)
	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Invalid target URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Custom Director to strip the prefix
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Request path: /projects/{name}/chat/proxy/{id}/foo/bar
		// Target path: /foo/bar

		// Find where /proxy/{id} ends
		prefix := fmt.Sprintf("/projects/%s/chat/proxy/%s", project.Name, chatID)
		if strings.HasPrefix(req.URL.Path, prefix) {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
		}
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}

		// Update headers
		req.Host = target.Host
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Header.Set("X-Forwarded-Proto", "http")
	}

	// Modify transport to ignore certificate errors (though we use http)
	// and handle WebSocket upgrades if needed (httputil usually handles this)

	isWS := strings.ToLower(r.Header.Get("Upgrade")) == "websocket"
	proxy.ServeHTTP(w, r)

	if isWS {
		log.Printf("WebSocket closed for chat %s, stopping session", chatID)
		StopChatSession(project.Name, chatID)
	}
}

func HandleChatDelete(w http.ResponseWriter, r *http.Request, project Project, chatID string) {
	permanent := r.URL.Query().Get("permanent") == "true"
	if permanent {
		PermanentlyDeleteChat(project.Name, chatID)
	} else {
		DeleteChat(project.Name, chatID)
	}
	StopChatSession(project.Name, chatID)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "chatListChanged")
		w.Header().Set("HX-Push-Url", fmt.Sprintf("/projects/%s/chat", project.Name))
		fmt.Fprint(w, "<div class='flex flex-col items-center justify-center h-full text-slate-500'>Select a chat or create a new one</div>")
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/projects/%s/chat", project.Name), http.StatusSeeOther)
}

func HandleChatRestore(w http.ResponseWriter, r *http.Request, project Project, chatID string) {
	RestoreChat(project.Name, chatID)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "chatListChanged")
		w.Header().Set("HX-Push-Url", fmt.Sprintf("/projects/%s/chat?id=%s", project.Name, chatID))
		HandleChatMessages(w, r, project, chatID, nil)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/projects/%s/chat?id=%s", project.Name, chatID), http.StatusSeeOther)
}

func HandleChatSend(w http.ResponseWriter, r *http.Request, project Project) {
	log.Printf("[DEBUG] HandleChatSend: project=%s", project.Name)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil && err != http.ErrNotMultipart {
		log.Printf("[ERROR] HandleChatSend: Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	chatID := r.FormValue("chat_id")
	content := r.FormValue("content")
	if chatID == "" || content == "" {
		http.Error(w, "Missing chat_id or content", http.StatusBadRequest)
		return
	}

	chat, err := GetChat(project.Name, chatID)
	if err != nil {
		http.Error(w, "Chat not found", http.StatusNotFound)
		return
	}

	var images []string
	if r.MultipartForm != nil {
		files := r.MultipartForm.File["images"]
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				continue
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				continue
			}

			// Convert to base64 data URL

			mimeType := fileHeader.Header.Get("Content-Type")
			base64Data := base64.StdEncoding.EncodeToString(data)
			images = append(images, fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data))
		}
	}

	userMsg := ChatMessage{
		Role:      "user",
		Content:   content,
		Images:    images,
		Timestamp: time.Now(),
	}

	chat.Messages = append(chat.Messages, userMsg)
	chat.UpdatedAt = time.Now()

	// Update title if it's the first message
	if len(chat.Messages) == 1 {
		title := content
		if len(title) > 30 {
			title = title[:27] + "..."
		}
		chat.Title = title
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Trigger", "chatListChanged")
		}
	}

	log.Printf("[DEBUG] HandleChatSend: project=%s, chat_id=%s, content=%s", project.Name, chatID, content)
	SaveChat(chat)

	// Start the chat session (this will ensure container is running)
	session, err := GetOrCreateChatSession(project, chatID)
	sessionError := ""
	if err != nil {
		log.Printf("[ERROR] HandleChatSend: Failed to start chat session: %v", err)
		sessionError = fmt.Sprintf("Failed to start chat session: %v", err)
	} else if session != nil && session.OpenCodeSessionID != "" {
		// Send message to OpenCode container asynchronously
		client := NewOpenCodeClient(session.Port)
		log.Printf("[DEBUG] HandleChatSend: Sending message to container: session_id=%s, port=%s", session.OpenCodeSessionID, session.Port)
		if err := client.SendMessageAsync(session.OpenCodeSessionID, content); err != nil {
			log.Printf("[ERROR] HandleChatSend: Failed to send async message: %v", err)
			sessionError = fmt.Sprintf("Failed to send message to AI: %v", err)
		}
	} else {
		log.Printf("[WARN] HandleChatSend: No active session or OpenCodeSessionID after GetOrCreateChatSession")
		sessionError = "Failed to establish a valid AI session."
	}

	// Trigger UI update - return both the confirmed user message AND the processing indicator
	tmpl, err := parseTemplate("api/templates/chat.html")
	if err != nil {
		log.Printf("[ERROR] HandleChatSend: Template error: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the user message first
	tmpl.ExecuteTemplate(w, "chat-message", map[string]interface{}{
		"Project":   project,
		"ChatID":    chatID,
		"Role":      userMsg.Role,
		"Timestamp": userMsg.Timestamp,
		"Content":   userMsg.Content,
		"Images":    userMsg.Images,
	})

	// Then the processing indicator or error
	if sessionError != "" {
		tmpl.ExecuteTemplate(w, "chat-error", map[string]interface{}{
			"Error": sessionError,
		})
	} else {
		data := map[string]interface{}{
			"Project":      project,
			"Current":      chat,
			"IsGenerating": true,
		}
		tmpl.ExecuteTemplate(w, "chat-processing", data)
	}
}

func HandleChatSync(w http.ResponseWriter, r *http.Request, project Project, chatID string) {
	log.Printf("[DEBUG] HandleChatSync: project=%s, chatID=%s", project.Name, chatID)
	// Start session if needed (it should be running if we just got a completion event)
	session, err := GetOrCreateChatSession(project, chatID)

	tmpl, tmplErr := parseTemplate("api/templates/chat.html")
	if tmplErr != nil {
		log.Printf("[ERROR] HandleChatSync: Template error: %v", tmplErr)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err != nil {
		log.Printf("[ERROR] HandleChatSync: Failed to connect to chat session: %v", err)
		tmpl.ExecuteTemplate(w, "chat-error", map[string]interface{}{
			"Error": fmt.Sprintf("Failed to connect to chat session: %v", err),
		})
		return
	}

	if err := SyncSession(session); err != nil {
		log.Printf("[ERROR] HandleChatSync: Failed to sync session: %v", err)
	}

	chat, err := GetChat(project.Name, chatID)
	if err != nil {
		log.Printf("[ERROR] HandleChatSync: Chat not found: %v", err)
		http.Error(w, "Chat not found", http.StatusNotFound)
		return
	}

	client := NewOpenCodeClient(session.Port)

	isGenerating := false
	if generating, err := client.IsGenerating(session.OpenCodeSessionID); err == nil {
		isGenerating = generating
	}
	log.Printf("[DEBUG] HandleChatSync: isGenerating=%v", isGenerating)

	data := map[string]interface{}{
		"Project":      project,
		"Current":      chat,
		"IsGenerating": isGenerating,
	}

	// If we are generating, or if the last message is from user (and we are waiting for assistant),
	// return the processing indicator to keep polling.
	if isGenerating || (len(chat.Messages) > 0 && chat.Messages[len(chat.Messages)-1].Role == "user") {
		tmpl.ExecuteTemplate(w, "chat-processing", data)
		return
	}

	// If finished generating, return the last assistant message to replace the processing indicator
	if len(chat.Messages) > 0 {
		lastMsg := chat.Messages[len(chat.Messages)-1]
		if lastMsg.Role == "assistant" {
			tmpl.ExecuteTemplate(w, "chat-message", map[string]interface{}{
				"Project":   project,
				"ChatID":    chatID,
				"Role":      lastMsg.Role,
				"Timestamp": lastMsg.Timestamp,
				"Content":   lastMsg.Content,
				"Images":    lastMsg.Images,
			})
			return
		}
	}

	// If no assistant message found and no longer generating, stop polling by returning nothing
	w.WriteHeader(http.StatusOK)
}

func HandleChatBuild(w http.ResponseWriter, r *http.Request, project Project) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		// Try to read body if form value is empty
		if err := r.ParseForm(); err == nil {
			content = r.Form.Get("content")
		}
	}

	if content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	target := fmt.Sprintf("/projects/%s/jobs?prompt=%s", project.Name, url.QueryEscape(content))
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", target)
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, target, http.StatusSeeOther)
}

func HandleChatRename(w http.ResponseWriter, r *http.Request, project Project, chatID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	chat, err := GetChat(project.Name, chatID)
	if err != nil {
		http.Error(w, "Chat not found", http.StatusNotFound)
		return
	}

	chat.Title = title
	chat.UpdatedAt = time.Now()
	if err := SaveChat(chat); err != nil {
		http.Error(w, "Failed to save chat", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "chatListChanged")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/projects/%s/chat?id=%s", project.Name, chatID), http.StatusSeeOther)
}
