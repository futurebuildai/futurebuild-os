package readiness

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Probe verifies S3/MinIO connectivity by checking if the configured bucket exists.
type S3Probe struct {
	endpoint  string
	bucket    string
	accessKey string
	secretKey string
}

// NewS3Probe creates a probe that connects to the S3-compatible endpoint
// and calls BucketExists on the configured bucket.
func NewS3Probe(endpoint, bucket, accessKey, secretKey string) *S3Probe {
	return &S3Probe{
		endpoint:  endpoint,
		bucket:    bucket,
		accessKey: accessKey,
		secretKey: secretKey,
	}
}

func (p *S3Probe) Name() string { return "s3" }

func (p *S3Probe) Check(ctx context.Context) CheckResult {
	start := time.Now()

	if p.endpoint == "" || p.accessKey == "" || p.secretKey == "" {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusNotConfigured,
			Message:  "S3_ENDPOINT, S3_ACCESS_KEY, or S3_SECRET_KEY not set",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	// Determine TLS based on endpoint scheme.
	useSSL := !strings.HasPrefix(p.endpoint, "http://")
	// Strip scheme for minio client (it expects host:port only).
	endpoint := p.endpoint
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(p.accessKey, p.secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("client creation failed: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	if p.bucket == "" {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusNotConfigured,
			Message:  "S3_BUCKET not set",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	exists, err := client.BucketExists(ctx, p.bucket)
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("bucket check failed: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	if !exists {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("bucket %q does not exist", p.bucket),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return CheckResult{
		Name:     p.Name(),
		Status:   StatusHealthy,
		Duration: time.Since(start).Milliseconds(),
	}
}
