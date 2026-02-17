package sync

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/dopejs/gozen/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const s3ObjectKey = "zen-sync.json"

// S3Backend implements Backend using S3-compatible storage via minio-go.
type S3Backend struct {
	client *minio.Client
	bucket string
}

// NewS3Backend creates an S3Backend from SyncConfig.
func NewS3Backend(cfg *config.SyncConfig) (*S3Backend, error) {
	endpoint := cfg.Endpoint
	useSSL := true
	// Strip scheme for minio client
	if len(endpoint) > 8 && endpoint[:8] == "https://" {
		endpoint = endpoint[8:]
	} else if len(endpoint) > 7 && endpoint[:7] == "http://" {
		endpoint = endpoint[7:]
		useSSL = false
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: useSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("s3 init: %w", err)
	}
	return &S3Backend{client: client, bucket: cfg.Bucket}, nil
}

func (b *S3Backend) Name() string { return "s3" }

func (b *S3Backend) Download(ctx context.Context) ([]byte, error) {
	obj, err := b.client.GetObject(ctx, b.bucket, s3ObjectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("s3 download: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		// Check if object doesn't exist
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return nil, nil
		}
		return nil, fmt.Errorf("s3 download: %w", err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data, nil
}

func (b *S3Backend) Upload(ctx context.Context, data []byte) error {
	reader := bytes.NewReader(data)
	_, err := b.client.PutObject(ctx, b.bucket, s3ObjectKey, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		return fmt.Errorf("s3 upload: %w", err)
	}
	return nil
}
