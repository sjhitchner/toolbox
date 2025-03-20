package fileutils

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Open returns an io.ReadCloser for the specified file path (GCS, local, S3).
func Open(ctx context.Context, filePath string) (io.ReadCloser, error) {
	u, err := url.Parse(filePath)
	if err != nil {
		if !strings.HasPrefix(filePath, "gs://") && !strings.HasPrefix(filePath, "s3://") {
			return readFileLocalToReader(filePath)
		}
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	switch u.Scheme {
	case "gs":
		return readFileGCSToReader(ctx, u)
	case "s3":
		return readFileS3ToReader(ctx, u)
	case "file":
		return readFileLocalToReader(u.Path)
	default:
		if u.Scheme == "" {
			return readFileLocalToReader(filePath)
		}
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
}

func readFileLocalToReader(filePath string) (io.ReadCloser, error) {
	return os.Open(filePath)
}

func readFileGCSToReader(ctx context.Context, u *url.URL) (io.ReadCloser, error) {
	bucketName := u.Host
	objectName := strings.TrimPrefix(u.Path, "/")

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	rc, err := client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		client.Close() // Close client if reader creation fails
		return nil, fmt.Errorf("failed to read GCS object: %w", err)
	}
	// Client is closed when rc is closed.
	return rc, nil
}

func readFileS3ToReader(ctx context.Context, u *url.URL) (io.ReadCloser, error) {
	bucket := u.Host
	key := strings.TrimPrefix(u.Path, "/")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 session: %w", err)
	}

	svc := s3.New(sess)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := svc.GetObjectWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object: %w", err)
	}

	return result.Body, nil // result.Body is already an io.ReadCloser
}

func CreateDirectory(path string) error {
	// Check if the directory already exists
	if _, err := os.Stat(path); err == nil {
		// Directory exists, no need to create
		return nil
	}

	// Create the directory and any necessary parent directories
	err := os.MkdirAll(path, os.ModePerm) // Use 0755 for more restrictive permissions
	if err != nil {
		return fmt.Errorf("failed to create directory path: %w", err) // Wrap the error
	}

	return nil
}
