package api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

type ChatMessage struct {
	Role      string    `json:"role" yaml:"role"` // "user" or "assistant"
	Content   string    `json:"content" yaml:"content"`
	Images    []string  `json:"images,omitempty" yaml:"images,omitempty"` // Base64 data URLs
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
}

type Chat struct {
	ID                string        `json:"id" yaml:"id"`
	OpenCodeSessionID string        `json:"opencode_session_id,omitempty" yaml:"opencode_session_id,omitempty"`
	Title             string        `json:"title" yaml:"title"`
	Project           string        `json:"project" yaml:"project"`
	Messages          []ChatMessage `json:"messages" yaml:"messages"`
	CreatedAt         time.Time     `json:"created_at" yaml:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at" yaml:"updated_at"`
	DeletedAt         *time.Time    `json:"deleted_at,omitempty" yaml:"deleted_at,omitempty"`
}

func GetChat(projectName, chatID string) (*Chat, error) {
	project, err := GetProject(projectName)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(project.ChatsPath(), chatID+".yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var chat Chat
	if err := yaml.Unmarshal(data, &chat); err != nil {
		return nil, err
	}
	return &chat, nil
}

func SaveChat(chat *Chat) error {
	project, err := GetProject(chat.Project)
	if err != nil {
		return err
	}
	path := project.ChatsPath()
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(chat)
	if err != nil {
		return err
	}

	// Atomic write using a temporary file
	tmpPath := filepath.Join(path, chat.ID+".yml.tmp")
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, filepath.Join(path, chat.ID+".yml"))
}

func ListChats(projectName string) ([]Chat, error) {
	project, err := GetProject(projectName)
	if err != nil {
		return nil, err
	}
	path := project.ChatsPath()
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Chat{}, nil
		}
		return nil, err
	}

	var chats []Chat
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yml" {
			chatID := entry.Name()[:len(entry.Name())-4]
			chat, err := GetChat(projectName, chatID)
			if err == nil && chat.DeletedAt == nil {
				chats = append(chats, *chat)
			}
		}
	}

	sort.Slice(chats, func(i, j int) bool {
		return chats[i].UpdatedAt.After(chats[j].UpdatedAt)
	})

	return chats, nil
}

func ListDeletedChats(projectName string) ([]Chat, error) {
	project, err := GetProject(projectName)
	if err != nil {
		return nil, err
	}
	path := project.ChatsPath()
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Chat{}, nil
		}
		return nil, err
	}

	var chats []Chat
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yml" {
			chatID := entry.Name()[:len(entry.Name())-4]
			chat, err := GetChat(projectName, chatID)
			if err == nil && chat.DeletedAt != nil {
				chats = append(chats, *chat)
			}
		}
	}

	sort.Slice(chats, func(i, j int) bool {
		return chats[i].DeletedAt.After(*chats[j].DeletedAt)
	})

	return chats, nil
}

func CleanupOldDeletedChats(projectName string) error {
	chats, err := ListDeletedChats(projectName)
	if err != nil {
		return err
	}

	sixtyDaysAgo := time.Now().AddDate(0, 0, -60)
	for _, chat := range chats {
		if chat.DeletedAt != nil && chat.DeletedAt.Before(sixtyDaysAgo) {
			if err := PermanentlyDeleteChat(projectName, chat.ID); err != nil {
				log.Printf("Failed to permanently delete old chat %s: %v", chat.ID, err)
			}
		}
	}
	return nil
}

func GenerateChatID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func DeleteChat(projectName, chatID string) error {
	chat, err := GetChat(projectName, chatID)
	if err != nil {
		return err
	}
	now := time.Now()
	chat.DeletedAt = &now
	return SaveChat(chat)
}

func RestoreChat(projectName, chatID string) error {
	chat, err := GetChat(projectName, chatID)
	if err != nil {
		return err
	}
	chat.DeletedAt = nil
	return SaveChat(chat)
}

func PermanentlyDeleteChat(projectName, chatID string) error {
	project, err := GetProject(projectName)
	if err != nil {
		return err
	}
	return os.Remove(filepath.Join(project.ChatsPath(), chatID+".yml"))
}

// For AI response, we'll need to call the AI.
// Since we don't have a direct LLM client here, we might need to use the job system or another way.
// But for now, let's implement the UI and basic storage.
