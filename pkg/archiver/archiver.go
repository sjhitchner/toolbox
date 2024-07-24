package archiver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/sjhitchner/toolbox/pkg/streaming"
)

type File[T any] struct {
	part  int
	Buf   *bytes.Buffer
	enc   *json.Encoder
	Count int
}

func NewFile[T any](part int) File[T] {
	buf := &bytes.Buffer{}
	return File[T]{
		part:  part,
		Buf:   buf,
		enc:   json.NewEncoder(buf),
		Count: 0,
	}
}

func (t File[T]) Key() string {
	return fmt.Sprintf("%04d.json", t.part)
}

func (t *File[T]) Add(obj T) error {
	if err := t.enc.Encode(obj); err != nil {
		return err
	}
	t.Count++
	return nil
}

type Saver[T any] interface {
	Save(start time.Time, in <-chan File[T]) <-chan error
}

type Loader[T any] interface {
	List(start *time.Time) (<-chan string, <-chan error)
	Load(path string) (io.ReadCloser, error)
}

type Archiver[T any] interface {
	Archive(start time.Time, ch <-chan T) <-chan error
	Retrieve(from *time.Time) (<-chan T, <-chan error, error)
}

type BaseArchiver[T any] struct {
	saver   Saver[T]
	loader  Loader[T]
	bufSize int
}

func (t *BaseArchiver[T]) Archive(start time.Time, in <-chan T) <-chan error {
	uploadCh, addErrs := t.startArchiver(in)
	errUploads := t.saver.Save(start, uploadCh)
	errCh := streaming.Merge[error](addErrs, errUploads)
	return errCh
}

func (t *BaseArchiver[T]) startArchiver(in <-chan T) (<-chan File[T], <-chan error) {
	out := make(chan File[T])
	errCh := make(chan error)

	go func() {
		defer close(errCh)
		defer close(out)

		var part int

		file := NewFile[T](part)

		for record := range in {
			if err := file.Add(record); err != nil {
				errCh <- err
				continue
			}

			if file.Count >= t.bufSize {
				out <- file
				part++
				file = NewFile[T](part)
			}
		}
		out <- file
	}()

	return out, errCh
}

func (t *BaseArchiver[T]) Retrieve(from *time.Time) (<-chan T, <-chan error, error) {
	out := make(chan T)
	errCh := make(chan error)

	filesCh, errListCh := t.loader.List(from)

	go func() {
		defer close(out)
		defer close(errCh)

		for file := range filesCh {
			if err := t.retrieve(file, out, errCh); err != nil {
				errCh <- err
			}
		}
	}()

	return out, streaming.Merge[error](errCh, errListCh), nil
}

func (t *BaseArchiver[T]) retrieve(filename string, out chan<- T, errCh chan<- error) error {
	reader, err := t.loader.Load(filename)
	if err != nil {
		return err
	}
	defer reader.Close()

	dec := json.NewDecoder(reader)

	for {
		var obj T
		if err := dec.Decode(&obj); err == io.EOF {
			break
		} else if err != nil {
			errCh <- err
		}

		out <- obj
	}

	return nil
}

func NewArchiver[T any](path, dataType string, bufSize int) (Archiver[T], error) {
	if strings.HasPrefix(path, "s3://") {
		u, err := url.Parse(path)
		if err != nil {
			return nil, fmt.Errorf("Invalid S3 Path %+v", err)
		}

		s3Client, err := GetS3Client()
		if err != nil {
			return nil, err
		}

		return NewS3Archiver[T](s3Client, u.Host, u.Path[1:], dataType, bufSize)
	}

	return NewFileArchiver[T](path, dataType, bufSize)
}
