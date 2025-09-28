package pubsub

import "context"

type PubSub interface {
	SendMessage(ctx context.Context, appID string, taskID string) error
	PullMessage(ctx context.Context) (string, error)
	RenewLock(ctx context.Context, appID string) error
	DeleteLock(ctx context.Context, appID string) error
}

type Cache interface {
	GetAppPid(ctx context.Context, appID string) (int, error)
	SetAppPid(ctx context.Context, appID string, pid int) error
}
