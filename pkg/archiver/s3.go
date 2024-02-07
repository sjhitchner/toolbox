package archiver

import (
	"context"
	"io"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

const (
	S3DateFormat = "2006/01/02"
)

func NewS3Archiver[T any](client *s3.Client, bucket, path string, bufSize int) (*BaseArchiver[T], error) {
	return &BaseArchiver[T]{
		bufSize: bufSize,
		saver: &S3Saver[T]{
			client: client,
			bucket: bucket,
			path:   path,
		},
		loader: &S3Loader[T]{
			client: client,
			bucket: bucket,
			path:   path,
		},
	}, nil
}

type S3Saver[T any] struct {
	client *s3.Client
	bucket string
	path   string
}

func (t *S3Saver[T]) Save(start time.Time, uploads <-chan File[T]) <-chan error {
	errCh := make(chan error)

	go func() {
		defer close(errCh)

		// Create an uploader
		uploader := manager.NewUploader(t.client)

		for file := range uploads {
			key := path.Join(t.path, start.Format(S3DateFormat), file.Key())

			input := &s3.PutObjectInput{
				Bucket: aws.String(t.bucket),
				Key:    aws.String(key),
				Body:   file.Buf,
			}

			log.Info().
				Str("key", key).
				Str("bucket", t.bucket).
				Msg("Started uploading archive to S3")

			ctx := context.Background()
			_, err := uploader.Upload(ctx, input)
			if err != nil {
				errCh <- err
			}

			log.Info().
				Str("key", key).
				Str("bucket", t.bucket).
				Msg("Finished uploading archive to S3")
		}
	}()

	return errCh
}

type S3Loader[T any] struct {
	client *s3.Client
	bucket string
	path   string
}

func (t *S3Loader[T]) List(start time.Time) (<-chan string, <-chan error) {

	path := path.Join(t.path, start.Format(S3DateFormat))

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(t.bucket),
		Prefix: aws.String(path),
	}

	out := make(chan string)
	errCh := make(chan error)

	go func() {
		defer close(out)
		defer close(errCh)

		ctx := context.Background()

		paginator := s3.NewListObjectsV2Paginator(t.client, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				errCh <- err
				continue
			}

			for _, obj := range page.Contents {
				filename := *obj.Key

				log.Debug().
					Str("bucket", t.bucket).
					Str("path", filename).
					Msg("Found S3 file")

				out <- filename
			}
		}

		log.Debug().
			Str("bucket", t.bucket).
			Str("path", path).
			Msg("Finished listing S3 files")
	}()

	return out, errCh
}

func (t *S3Loader[T]) Load(path string) (io.ReadCloser, error) {
	ctx := context.Background()

	input := &s3.GetObjectInput{
		Bucket: aws.String(t.bucket),
		Key:    aws.String(path),
	}

	output, err := t.client.GetObject(ctx, input)
	if err != nil {
		return nil, err
	}

	//defer output.Body.Close()
	return output.Body, nil
}
