package api

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ChatSession struct {
	ChatID            string
	OpenCodeSessionID string
	ProjectName       string
	Port              string
	ContainerID       string
	LastActive        time.Time
}

var (
	sessionsMutex sync.RWMutex
	sessions      = make(map[string]*ChatSession) // Key: ProjectName + ":" + ChatID

	warmingMutex sync.Mutex
	warming      = make(map[string]chan struct{}) // Key: ProjectName + ":" + ChatID
)

func WarmChatSession(project Project, chatID string) {
	key := project.Name + ":" + chatID

	sessionsMutex.RLock()
	session, exists := sessions[key]
	sessionsMutex.RUnlock()
	if exists && session != nil {
		// Check health of existing session
		client := NewOpenCodeClient(session.Port)
		if err := client.Health(); err == nil {
			return
		}
		log.Printf("[DEBUG] WarmChatSession: Session %s exists but is unhealthy, warming to restart...", key)
	}

	warmingMutex.Lock()
	if _, beingWarmed := warming[key]; beingWarmed {
		warmingMutex.Unlock()
		return
	}

	waitCh := make(chan struct{})
	warming[key] = waitCh
	warmingMutex.Unlock()

	log.Printf("[DEBUG] Warming chat session for %s...", key)
	go func() {
		defer func() {
			warmingMutex.Lock()
			delete(warming, key)
			close(waitCh)
			warmingMutex.Unlock()
		}()

		_, _ = getOrCreateChatSession(project, chatID, false)
	}()
}

func init() {
	DiscoverSessions()
	go startReaper()
}

func DiscoverSessions() {
	log.Printf("Discovering existing chat sessions...")
	// Format: name|labels
	cmdExec := exec.Command("podman", "ps", "--format", "{{.Names}}|{{.Labels}}", "--filter", "label=type=chat")
	out, err := cmdExec.Output()
	if err != nil {
		// This is expected if podman is not installed or no containers exist
		return
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		labels := parts[1]

		projectName := ""
		chatID := ""

		for _, label := range strings.Split(labels, ",") {
			kv := strings.Split(label, "=")
			if len(kv) == 2 {
				if kv[0] == "project" {
					projectName = kv[1]
				} else if kv[0] == "chat_id" {
					chatID = kv[1]
				}
			}
		}

		// Fallback to name parsing if labels are missing chat_id (for older containers)
		if chatID == "" && strings.HasPrefix(name, "overdrive-chat-") {
			nameParts := strings.Split(name, "-")
			if len(nameParts) >= 4 {
				// overdrive-chat-project-chatID
				// This is risky if project has hyphens, but it's a fallback
				chatID = nameParts[len(nameParts)-1]
				if projectName == "" {
					projectName = strings.Join(nameParts[2:len(nameParts)-1], "-")
				}
			}
		}

		if projectName != "" && chatID != "" {
			key := projectName + ":" + chatID
			log.Printf("Found existing chat container: %s (project: %s, chat: %s)", name, projectName, chatID)

			// Get port
			portCmd := exec.Command("podman", "port", name, "3000")
			portOut, err := portCmd.Output()
			if err != nil {
				log.Printf("Failed to get port for %s: %v", name, err)
				continue
			}

			// Output format: 3000/tcp -> 0.0.0.0:38475
			portMapping := string(portOut)
			portParts := strings.Split(portMapping, ":")
			if len(portParts) >= 2 {
				port := strings.TrimSpace(portParts[len(portParts)-1])

				session := &ChatSession{
					ChatID:      chatID,
					ProjectName: projectName,
					Port:        port,
					LastActive:  time.Now(),
				}

				sessionsMutex.Lock()
				sessions[key] = session
				sessionsMutex.Unlock()
				log.Printf("Adopted session %s on port %s", key, port)
			}
		}
	}
}

func startReaper() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		sessionsMutex.Lock()
		for key, session := range sessions {
			if time.Since(session.LastActive) > 5*time.Minute {
				log.Printf("Reaping idle chat session %s (last active: %v)", key, session.LastActive)
				go StopChatSession(session.ProjectName, session.ChatID)
			}
		}
		sessionsMutex.Unlock()
	}
}

func GetChatSession(projectName, chatID string) *ChatSession {
	sessionsMutex.RLock()
	defer sessionsMutex.RUnlock()
	return sessions[projectName+":"+chatID]
}

func GetOrCreateChatSession(project Project, chatID string) (*ChatSession, error) {
	return getOrCreateChatSession(project, chatID, true)
}

func getOrCreateChatSession(project Project, chatID string, checkWarming bool) (*ChatSession, error) {
	key := project.Name + ":" + chatID

	if checkWarming {
		warmingMutex.Lock()
		for {
			if waitCh, beingWarmed := warming[key]; beingWarmed {
				warmingMutex.Unlock()
				log.Printf("[DEBUG] GetOrCreateChatSession: Session %s is being warmed, waiting...", key)
				<-waitCh
				warmingMutex.Lock()
				continue
			}
			break
		}

		// Check if it exists and is healthy
		sessionsMutex.RLock()
		session, exists := sessions[key]
		sessionsMutex.RUnlock()

		if exists {
			log.Printf("[DEBUG] GetOrCreateChatSession: session exists for %s, port=%s. Checking health...", key, session.Port)
			client := NewOpenCodeClient(session.Port)
			if err := client.Health(); err == nil {
				session.LastActive = time.Now()
				warmingMutex.Unlock()
				return session, nil
			}
			log.Printf("[DEBUG] GetOrCreateChatSession: session %s is unhealthy, will stop and restart", key)
		}

		// Set warming flag
		waitCh := make(chan struct{})
		warming[key] = waitCh
		warmingMutex.Unlock()

		defer func() {
			warmingMutex.Lock()
			delete(warming, key)
			close(waitCh)
			warmingMutex.Unlock()
		}()
	}

	// Re-check existence and health (either we held the warming flag or checkWarming was false)
	sessionsMutex.RLock()
	session, exists := sessions[key]
	sessionsMutex.RUnlock()

	if exists {
		client := NewOpenCodeClient(session.Port)
		if err := client.Health(); err == nil {
			session.LastActive = time.Now()
			return session, nil
		}
		log.Printf("[DEBUG] GetOrCreateChatSession: stopping unhealthy session %s", key)
		StopChatSession(project.Name, chatID)
		// Fall through to start a new session
	}

	// Ensure OpenCode session exists
	// We need to load the Chat to check/save the OpenCodeSessionID
	chat, err := GetChat(project.Name, chatID)
	if err != nil {
		log.Printf("[ERROR] GetOrCreateChatSession: Failed to load chat: %v", err)
		return nil, fmt.Errorf("failed to load chat: %v", err)
	}

	opencodeSessionID := chat.OpenCodeSessionID

	// Check if session data exists in volume
	dataDir := filepath.Join("projects", project.Name, "chats_data", chatID)
	dbPath := filepath.Join(dataDir, "opencode.db")
	hasData := false
	if _, err := os.Stat(dbPath); err == nil {
		hasData = true
		log.Printf("[DEBUG] GetOrCreateChatSession: Found existing session data in volume for %s", chatID)
	}

	log.Printf("[DEBUG] GetOrCreateChatSession: Starting chat server for %s", key)
	if err := CheckConcurrency(project, "chat"); err != nil {
		log.Printf("[WARN] GetOrCreateChatSession: Concurrency limit reached for %s: %v", key, err)
		return nil, fmt.Errorf("concurrency limit reached: %v", err)
	}
	port, err := startChatServer(project, chatID, opencodeSessionID)
	if err != nil {
		log.Printf("[ERROR] GetOrCreateChatSession: Failed to start chat server: %v", err)
		return nil, err
	}

	client := NewOpenCodeClient(port)

	// Poll for health (up to 2 minutes)
	log.Printf("[DEBUG] GetOrCreateChatSession: Waiting for health on port %s", port)
	healthy := false
	// Use shorter interval for faster startup detection
	for i := 0; i < 1200; i++ { // 120 seconds with 100ms polling
		if err := client.Health(); err == nil {
			healthy = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !healthy {
		log.Printf("[ERROR] GetOrCreateChatSession: Chat server failed to become healthy on port %s", port)
		// Kill container if it failed to start properly
		StopChatSession(project.Name, chatID)
		return nil, fmt.Errorf("chat server failed to become healthy")
	}

	// Check if the session actually exists in the container
	sessionValid := false
	if opencodeSessionID != "" {
		// Try a few times to get messages, as it might be loading from DB
		for i := 0; i < 5; i++ {
			if _, err := client.GetSessionMessages(opencodeSessionID); err == nil {
				log.Printf("[DEBUG] GetOrCreateChatSession: Validated existing OpenCode session: %s", opencodeSessionID)
				sessionValid = true
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		if !sessionValid && hasData {
			// If we have data, we assume the session IS valid even if we can't get messages yet.
			// This prevents creating a new session and orphaning old data.
			log.Printf("[DEBUG] GetOrCreateChatSession: Session %s invalid in container but exists in volume, assuming valid", opencodeSessionID)
			sessionValid = true
		}
	}

	if !sessionValid {
		log.Printf("[DEBUG] GetOrCreateChatSession: Creating new OpenCode session")
		// Create new session in OpenCode
		title := chat.Title
		if title == "" {
			title = "New Chat"
		}

		// Create session
		newSessionID, err := client.CreateSession(title)
		if err == nil {
			opencodeSessionID = newSessionID
			log.Printf("[DEBUG] GetOrCreateChatSession: Created new OpenCode session: %s", opencodeSessionID)
			// Save to Chat
			chat.OpenCodeSessionID = opencodeSessionID
			SaveChat(chat)
		} else {
			log.Printf("[ERROR] GetOrCreateChatSession: Failed to create session: %v", err)
		}

		if opencodeSessionID == "" {
			log.Printf("[WARN] Failed to create OpenCode session for chat %s", chatID)
		}
	}

	session = &ChatSession{
		ChatID:            chatID,
		OpenCodeSessionID: opencodeSessionID,
		ProjectName:       project.Name,
		Port:              strings.TrimSpace(port),
		LastActive:        time.Now(),
	}

	sessionsMutex.Lock()
	sessions[key] = session
	sessionsMutex.Unlock()

	return session, nil
}

func SyncSession(session *ChatSession) error {
	if session == nil || session.OpenCodeSessionID == "" {
		return nil
	}

	// Use a shorter timeout for syncing to avoid stalling
	client := NewOpenCodeClientWithTimeout(session.Port, 5*time.Second)

	messages, err := client.GetSessionMessages(session.OpenCodeSessionID)
	if err != nil {
		return fmt.Errorf("failed to get messages: %v", err)
	}

	chat, err := GetChat(session.ProjectName, session.ChatID)
	if err != nil {
		return fmt.Errorf("failed to load chat: %v", err)
	}

	updated := false
	if len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		if lastMsg.Role == "assistant" {
			shouldAppend := true
			if len(chat.Messages) > 0 {
				lastLocalMsg := &chat.Messages[len(chat.Messages)-1]
				if lastLocalMsg.Role == "assistant" {
					if lastLocalMsg.Content != lastMsg.Content {
						lastLocalMsg.Content = lastMsg.Content
						updated = true
					}
					shouldAppend = false
				}
			}

			if shouldAppend {
				chat.Messages = append(chat.Messages, ChatMessage{
					Role:      "assistant",
					Content:   lastMsg.Content,
					Timestamp: time.Now(),
				})
				updated = true
			}
		}
	}

	if updated {
		chat.UpdatedAt = time.Now()
		return SaveChat(chat)
	}
	return nil
}

func startChatServer(project Project, chatID string, opencodeSessionID string) (string, error) {
	cmd := exec.Command("./scripts/chat-server")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "CHAT_ID="+chatID)
	cmd.Env = append(cmd.Env, "PROJECT_NAME="+project.Name)

	cmd.Env = append(cmd.Env, "REPO_URL="+project.RepoURL)
	cmd.Env = append(cmd.Env, "PRIMARY_BRANCH="+project.PrimaryBranch)

	// Use new LLM config
	cmd.Env = append(cmd.Env, "LLM_PROVIDER="+project.Chat.LLM.Provider)
	cmd.Env = append(cmd.Env, "MODEL="+project.Chat.LLM.Model)
	cmd.Env = append(cmd.Env, "HARNESS_PROVIDER="+project.Chat.Harness.Provider)
	cmd.Env = append(cmd.Env, "AUTH_JSON="+project.GenerateAuthJSON(project.Chat))

	if opencodeSessionID != "" {
		cmd.Env = append(cmd.Env, "OPENCODE_SESSION_ID="+opencodeSessionID)
	}

	// SSH keys might be needed for private repos
	if project.SSHKey != "" {
		if filepath.IsAbs(project.SSHKey) {
			cmd.Env = append(cmd.Env, "SSH_KEY="+project.SSHKey)
		} else {
			cmd.Env = append(cmd.Env, "SSH_KEY="+project.SSHKey)
		}
	}
	if project.SSHConfig != "" {
		cmd.Env = append(cmd.Env, "SSH_CONFIG="+project.SSHConfig)
	}

	log.Printf("Starting chat server for %s/%s...", project.Name, chatID)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("failed to start chat server: %s, stderr: %s", err, string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to start chat server: %v", err)
	}

	port := string(out)
	log.Printf("Chat server started on port %s", port)
	return port, nil
}

func StopChatSession(projectName, chatID string) {
	key := projectName + ":" + chatID
	sessionsMutex.Lock()
	session, exists := sessions[key]
	if exists {
		delete(sessions, key)
	}
	sessionsMutex.Unlock()

	// Even if it wasn't in the map, we want to make sure the container is killed
	log.Printf("Stopping chat session %s (exists=%v)", key, exists)

	if exists && session != nil {
		log.Printf("[DEBUG] StopChatSession: Syncing session %s before stopping...", key)
		// Sync before killing. The client now has a 10s timeout.
		if err := SyncSession(session); err != nil {
			log.Printf("[WARN] Failed to sync session %s before stopping: %v", key, err)
		}
	}

	// Kill the container
	containerName := fmt.Sprintf("overdrive-chat-%s-%s", projectName, chatID)
	// Use 'rm -f' instead of 'kill' to ensure it's removed even if stopped
	cmd := exec.Command("podman", "rm", "-f", containerName)
	if err := cmd.Run(); err != nil {
		// Only log error if it's not "no such container"
		log.Printf("Note: podman rm -f %s: %v", containerName, err)
	}
}
