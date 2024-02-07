package archiver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/sjhitchner/toolbox/pkg/streaming"
)

type File[T any] struct {
	prefix string
	part   int
	Buf    *bytes.Buffer
	enc    *json.Encoder
	Count  int
}

func NewFile[T any](prefix string, part int) File[T] {
	buf := &bytes.Buffer{}
	return File[T]{
		prefix: prefix,
		part:   part,
		Buf:    buf,
		enc:    json.NewEncoder(buf),
		Count:  0,
	}
}

func (t File[T]) Key() string {
	return filepath.Join(t.prefix, fmt.Sprintf("%04d.json", t.part))
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
	List(start time.Time) (<-chan string, <-chan error)
	Load(path string) (io.ReadCloser, error)
}

type BaseArchiver[T any] struct {
	saver   Saver[T]
	loader  Loader[T]
	bufSize int
}

func (t *BaseArchiver[T]) Archive(start time.Time, toAdd <-chan T, toDelete <-chan T) <-chan error {

	addCh, addErrs := t.startArchiver("add", toAdd)
	deleteCh, deleteErrs := t.startArchiver("delete", toDelete)

	uploadCh := streaming.Merge[File[T]](addCh, deleteCh)
	errUploads := t.saver.Save(start, uploadCh)

	errCh := streaming.Merge[error](addErrs, deleteErrs, errUploads)
	return errCh
}

func (t *BaseArchiver[T]) startArchiver(prefix string, in <-chan T) (<-chan File[T], <-chan error) {
	out := make(chan File[T])
	errCh := make(chan error)

	go func() {
		defer close(errCh)
		defer close(out)

		var part int

		file := NewFile[T](prefix, part)

		for record := range in {
			if err := file.Add(record); err != nil {
				errCh <- err
				continue
			}

			if file.Count >= t.bufSize {
				out <- file
				part++
				file = NewFile[T](prefix, part)
			}
		}
		out <- file
	}()

	return out, errCh
}

func (t *BaseArchiver[T]) Retrieve(from time.Time) (<-chan T, <-chan T, <-chan error, error) {
	toAdd := make(chan T)
	toDelete := make(chan T)
	errCh := make(chan error)

	filesCh, errListCh := t.loader.List(from)

	go func() {
		defer close(toAdd)
		defer close(toDelete)
		defer close(errCh)

		for file := range filesCh {
			if strings.HasSuffix(filepath.Dir(file), "add") {
				if err := t.retrieve(file, toAdd, errCh); err != nil {
					errCh <- err
				}

			} else if strings.HasSuffix(filepath.Dir(file), "delete") {
				if err := t.retrieve(file, toDelete, errCh); err != nil {
					errCh <- err
				}

			} else {
				errCh <- fmt.Errorf("invalid file type: %s", file)
			}
		}
	}()

	return toAdd, toDelete, streaming.Merge[error](errCh, errListCh), nil
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
