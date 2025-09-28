package storage

import (
	"context"
	"fmt"
	"os"
	"umami/pkg/utils"

	"cloud.google.com/go/storage"
)

type gcs struct {
	client *storage.Client
}

func NewGCS(ctx context.Context) (*gcs, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &gcs{
		client: client,
	}, nil
}

func (g *gcs) CreateBucket(ctx context.Context, name string) error {
	bucketName := fmt.Sprintf("umami-bucket-%s", utils.GetName(name))
	err := g.client.Bucket(bucketName).Create(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"), &storage.BucketAttrs{
		Location: "us-central1",
	})
	return err
}
