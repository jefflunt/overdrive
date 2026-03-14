package api

import (
	"context"
)

type HarnessSession interface {
	GetID() string
	GetPort() string
	Sync() error
}

type HarnessJob interface {
	GetID() string
	Wait(ctx context.Context) error
}

type Harness interface {
	StartChat(ctx context.Context, project Project, chatID string) (HarnessSession, error)
	RunBuild(ctx context.Context, project Project, jobID string, req JobRequest) (HarnessJob, error)
	StopChat(project Project, chatID string) error
}
