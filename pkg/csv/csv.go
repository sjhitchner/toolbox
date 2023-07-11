package csv

import (
	"encoding/csv"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	zl "github.com/rs/zerolog"
)

const (
	Comma = ','
	Tab   = '\t'
)

func StreamCSV(reader io.Reader, header bool, logger *zl.Logger) (<-chan []string, error) {
	return Stream(reader, header, Comma, logger)
}

func StreamTSV(reader io.Reader, header bool, logger *zl.Logger) (<-chan []string, error) {
	return Stream(reader, header, Tab, logger)
}

// Stream streams a csv file row by row attempts to close the reader if possible
func Stream(reader io.Reader, header bool, terminator rune, logger *zl.Logger) (<-chan []string, error) {
	out := make(chan []string)

	go func() {
		defer close(out)
		if rr, ok := reader.(io.ReadCloser); ok {
			defer rr.Close()
		}

		r := csv.NewReader(reader)
		r.Comma = terminator

		for i := 0; ; i++ {
			row, err := r.Read()
			if err == io.EOF {
				break
			}

			if err != nil {
				logger.Error().
					Int("line", i).
					Err(err).
					Strs("rows", row).
					Msg("failed reading line")
				continue
			}

			if header && i == 0 {
				continue
			}

			out <- row
		}
	}()

	return out, nil
}

func StreamDirectory(path string, header bool, logger *zl.Logger) (<-chan []string, error) {

	out := make(chan []string)

	files, err := WalkDirectory(path, logger)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	go func() {
		defer close(out)

		for file := range files {
			f, err := os.Open(file)
			if err != nil {
				logger.Error().
					Err(err).
					Str("file", file).
					Msg("failed opening file")
				continue
			}

			wg.Add(1)
			go func(filename string) {
				defer wg.Done()

				ch, err := Stream(f, header, Comma, logger)
				if err != nil {
					logger.Error().
						Err(err).
						Str("file", filename).
						Msg("failed streaming file")
					return
				}

				for r := range ch {
					out <- r
				}
			}(file)
		}
	}()

	return out, nil
}

// WalkDirectory looking for CSV files as defined by IsCSV function.
// Use WalkDirectoryFilter if you need to specify a custom extension.
func WalkDirectory(directory string, logger *zl.Logger) (<-chan string, error) {
	return WalkDirectoryFilter(directory, IsCSV, logger)
}

// FilterFunc
type FilterFunc func(string) bool

// WalkDirectoryFilter walks directory looking for files that match the FilterFunc
func WalkDirectoryFilter(directory string, filter FilterFunc, logger *zl.Logger) (<-chan string, error) {

	out := make(chan string)

	go func() {
		defer close(out)

		if err := filepath.WalkDir(directory, func(filename string, d fs.DirEntry, err error) error {

			if d.IsDir() {
				return nil
			}

			if !filter(filename) {
				return nil
			}

			out <- filename

			return nil

		}); err != nil {
			logger.Error().
				Err(err).
				Str("directory", directory).
				Msg("failed walking CSVs")
		}

	}()

	return out, nil
}

// IsCSV FilterFunc implementation
func IsCSV(filename string) bool {
	ext := filepath.Ext(filename)
	if ext != ".txt" && ext != ".csv" {
		return false
	}
	return true

}

// IsTSV FilterFunc implementation
func IsTSV(filename string) bool {
	ext := filepath.Ext(filename)
	if ext != ".txt" && ext != ".tsv" {
		return false
	}
	return true

}
