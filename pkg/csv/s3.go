/*
Package to interact with CSV files, finding, reading, processing.  Uses pipeline methodoligy to implement funcations that can be composed together to create complex pipelines
*/
package csv

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	zl "github.com/rs/zerolog"
	"github.com/sjhitchner/toolbox/pkg/streaming"
)

// PipelineFunc process a row to add information based on filename
type PipelineFunc func(filename string, row []string) []string

type S3Reader struct {
	sess          *session.Session
	s3            *s3.S3
	downloader    *s3manager.Downloader
	terminator    rune
	fileFilter    FilterFunc
	pipelineFuncs []PipelineFunc
	logger        zl.Logger
}

func NewS3Reader(sess *session.Session, logger *zl.Logger) *S3Reader {
	return &S3Reader{
		sess:       sess,
		downloader: s3manager.NewDownloader(sess),
		s3:         s3.New(sess),
		terminator: ',',
		fileFilter: IsCSV,
		logger:     logger.With().Logger(),
	}
}

func (t *S3Reader) WithFilter(fileFilter FilterFunc) *S3Reader {
	t.fileFilter = fileFilter
	return t
}

func (t *S3Reader) WithTerminator(terminator rune) *S3Reader {
	t.terminator = terminator
	return t
}

func (t *S3Reader) WithPipeline(fns ...PipelineFunc) *S3Reader {
	t.pipelineFuncs = fns
	return t
}

// listS3
// List files at S3 path
func (t *S3Reader) List(bucket, prefix string) (<-chan string, error) {

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	fmt.Println(input)

	result, err := t.s3.ListObjectsV2(input)
	if err != nil {
		return nil, err
	}

	fmt.Println(result)

	out := make(chan string)
	go func() {
		defer close(out)

		for _, obj := range result.Contents {
			file := aws.StringValue(obj.Key)

			t.logger.Info().
				Interface("result", result).
				Str("file", file).
				Msg("List S3")

			if !t.fileFilter(file) {
				continue
			}

			t.logger.Info().
				Str("file", file).
				Msg("adding file to queue")

			out <- file
		}
	}()

	return out, nil
}

// Stream - streams CSV from S3 row by row
func (t *S3Reader) Stream(bucket, path string, header bool) (<-chan []string, error) {
	logger := t.logger.With().
		Str("bucket", bucket).
		Str("path", path).
		Logger()

	logger.Info().Msg("Reading S3 Path")

	files, err := t.List(bucket, path)
	if err != nil {
		return nil, err
	}

	rows, err := t.stream(bucket, header, files)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (t *S3Reader) stream(bucket string, header bool, files <-chan string) (<-chan []string, error) {

	out := make(chan []string)
	var wg sync.WaitGroup

	downloadFn := func(filename string) {
		defer wg.Done()
		rows, err := t.download(bucket, filename, header)
		if err != nil {
			t.logger.Error().
				Err(err).
				Str("bucket", bucket).
				Str("key", filename).
				Msg("failed reading S3 file")
			return
		}

		for row := range rows {
			for _, fn := range t.pipelineFuncs {
				row = fn(filename, row)
			}
			out <- row
		}
	}

	go func() {
		defer close(out)

		for file := range files {
			wg.Add(1)
			go downloadFn(file)
		}

		wg.Wait()
	}()

	return out, nil
}

func (t *S3Reader) download(bucket, file string, header bool) (<-chan []string, error) {

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file),
	}

	t.logger.Info().
		Str("file", file).
		Msg("downloading s3 file")

	// buf := make([]byte, 100000000)
	writer := aws.NewWriteAtBuffer([]byte{})

	_, err := t.downloader.Download(writer, input)
	if err != nil {
		return nil, err
	}

	reader := NewReader(bytes.NewReader(writer.Bytes()))
	reader.Comma = t.terminator

	outCh, errCh, err := Stream[[]string](reader, header)
	if err != nil {
		return nil, err
	}
	return outCh, streaming.Error(errCh)
}

func ParseS3Path(s3Path string) (string, string, error) {
	u, err := url.Parse(s3Path)
	if err != nil {
		return "", "", err
	}

	return strings.TrimSpace(u.Host),
		strings.TrimSpace(u.Path[1:]), nil
}

/*

// Create a file to write the S3 Object contents to.
f, err := os.Create(filename)
if err != nil {
    return fmt.Errorf("failed to create file %q, %v", filename, err)
}

// Write the contents of S3 Object to the file
n, err := downloader.Download(f, &s3.GetObjectInput{
    Bucket: aws.String(myBucket),
    Key:    aws.String(myString),
})
if err != nil {
    return fmt.Errorf("failed to download file, %v", err)
}
fmt.Printf("file downloaded, %d bytes\n", n)
}

// readRows
// Based on provided csvPath the correct streaming method will be used
// - file
// - directory
// - s3 path (s3://)
func readRows(sess *session.Session, header bool, listID, csvPath string, logger *zl.Logger) (<-chan []string, error) {

	log := logger.With().
		Str("path", csvPath).
		Logger()

	if strings.HasPrefix(csvPath, "s3://") {
		return readCSVS3(sess, csvPath, header, &log)
	}

	fileInfo, err := os.Stat(csvPath)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return walkCSVs(csvPath, header, &log)
	}

	if listID == "" {
		return nil, fmt.Errorf("list-id must be specified")
	}

	return readCSVFile(listID, csvPath, header, &log)
}
*/
