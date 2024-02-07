package archiver

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sjhitchner/toolbox/pkg/files"

	"github.com/rs/zerolog/log"
)

const (
	FileDateFormat = "2006/01/02"
)

func NewFileArchiver[T any](archiveDir, dataType string, bufSize int) (*BaseArchiver[T], error) {

	path := filepath.Join(archiveDir, dataType)

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, err
	}

	return &BaseArchiver[T]{
		bufSize: bufSize,
		saver: &FileSaver[T]{
			path: path,
		},
		loader: &FileLoader[T]{
			path: path,
		},
	}, nil
}

type FileSaver[T any] struct {
	path string
}

func (t *FileSaver[T]) Save(start time.Time, uploads <-chan File[T]) <-chan error {
	errCh := make(chan error)

	go func() {
		defer close(errCh)

		for file := range uploads {
			filename := filepath.Join(t.path, start.Format(FileDateFormat), file.Key())

			if err := writeFile(filename, file.Buf); err != nil {
				errCh <- err
			}

			log.Info().
				Str("filename", filename).
				Msg("Finished writing archive")
		}
	}()

	return errCh
}

func writeFile(filename string, buf io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, buf); err != nil {
		return err
	}

	return nil
}

type FileLoader[T any] struct {
	path string
}

func (t *FileLoader[T]) List(start time.Time) (<-chan string, <-chan error) {
	path := filepath.Join(t.path, start.Format(FileDateFormat))
	return files.FindFiles(path), nil
}

func (t *FileLoader[T]) Load(filename string) (io.ReadCloser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return file, nil
}
