package api

import (
	"context"
	"fmt"
)

// OpenCodeHarness

type OpenCodeHarness struct{}

func (h *OpenCodeHarness) StartChat(ctx context.Context, project Project, chatID string) (HarnessSession, error) {
	// Re-use logic from GetOrCreateChatSession but adapted
	session, err := GetOrCreateChatSession(project, chatID)
	if err != nil {
		return nil, err
	}
	return &OpenCodeSessionAdapter{session: session}, nil
}

func (h *OpenCodeHarness) RunBuild(ctx context.Context, project Project, jobID string, req JobRequest) (HarnessJob, error) {
	// For now, OpenCodeHarness's RunBuild is handled by the worker's schedule/processJob loop
	// but we can wrap it if needed. The prompt says "spawns different containers/commands".
	// The current processJob does exactly this by calling scripts/work.
	return &OpenCodeJobAdapter{jobID: jobID}, nil
}

func (h *OpenCodeHarness) StopChat(project Project, chatID string) error {
	StopChatSession(project.Name, chatID)
	return nil
}

type OpenCodeSessionAdapter struct {
	session *ChatSession
}

func (s *OpenCodeSessionAdapter) GetID() string   { return s.session.ChatID }
func (s *OpenCodeSessionAdapter) GetPort() string { return s.session.Port }
func (s *OpenCodeSessionAdapter) Sync() error     { return SyncSession(s.session) }

type OpenCodeJobAdapter struct {
	jobID string
}

func (j *OpenCodeJobAdapter) GetID() string { return j.jobID }
func (j *OpenCodeJobAdapter) Wait(ctx context.Context) error {
	// Current worker system handles waiting via processJob.
	// This is a bit of a mismatch with the existing async job system.
	return nil
}

// ClaudeCodeHarness

type ClaudeCodeHarness struct{}

func (h *ClaudeCodeHarness) StartChat(ctx context.Context, project Project, chatID string) (HarnessSession, error) {
	return nil, fmt.Errorf("Claude Code harness not yet fully implemented")
}

func (h *ClaudeCodeHarness) RunBuild(ctx context.Context, project Project, jobID string, req JobRequest) (HarnessJob, error) {
	return nil, fmt.Errorf("Claude Code harness not yet fully implemented")
}

func (h *ClaudeCodeHarness) StopChat(project Project, chatID string) error {
	return nil
}

// Global factory

func GetHarness(provider string) Harness {
	switch provider {
	case "claude-code":
		return &ClaudeCodeHarness{}
	case "opencode":
		fallthrough
	default:
		return &OpenCodeHarness{}
	}
}
