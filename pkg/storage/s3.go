package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Service interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	GetSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

type S3Client struct {
	client *minio.Client
	bucket string
}

func NewS3Client(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*S3Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client: %w", err)
	}

	return &S3Client{
		client: minioClient,
		bucket: bucket,
	}, nil
}

func (s *S3Client) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	// Check if bucket exists, if not create it (auto-provisioning for dev convenience)
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	_, err = s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	return nil
}

func (s *S3Client) GetSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to sign url: %w", err)
	}
	return url.String(), nil
}
