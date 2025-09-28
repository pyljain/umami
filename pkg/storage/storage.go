package storage

import "context"

type Storage interface {
	CreateBucket(ctx context.Context, name string) error
}
