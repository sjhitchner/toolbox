package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
	zl "github.com/rs/zerolog"
	"github.com/sjhitchner/toolbox/pkg/streaming"
)

const (
	Comma = ','
	Tab   = '\t'

	CSVTag = "csv"
)

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

type Reader[T any] struct {
	Comma      rune
	DateFormat string
	HasHeader  bool
}

func NewReader[T any]() *Reader[T] {
	return &Reader[T]{
		Comma:      ',',
		DateFormat: defaultDateFormat,
		HasHeader:  false,
	}
}

func (t Reader[T]) Walk(path string) (<-chan T, <-chan error) {

	fileCh := make(chan string)
	outCh := make(chan T)
	errCh := make(chan error)
	readerErrCh := make(chan error)

	go func() {
		defer close(fileCh)
		defer close(readerErrCh)

		if err := filepath.WalkDir(path, func(filename string, d fs.DirEntry, err error) error {

			if d.IsDir() {
				return nil
			}

			fileCh <- filename

			return nil

		}); err != nil {
			readerErrCh <- err
		}
	}()

	go func() {
		defer close(outCh)
		defer close(errCh)

		for filename := range fileCh {
			reader, err := os.Open(filename)
			if err != nil {
				errCh <- err
				continue
			}

			ch, eCh, err := t.Stream(reader)
			if err != nil {
				errCh <- err
				continue
			}

			streaming.Copy(outCh, ch)
			streaming.Copy(errCh, eCh)
		}
	}()

	return outCh, streaming.Merge(errCh, readerErrCh)
}

// Stream reads a CSV file and returns a channel of the generic type T representing each row,
// and a channel for errors.
func (t *Reader[T]) Stream(reader io.Reader) (<-chan T, <-chan error, error) {
	outCh := make(chan T)
	errCh := make(chan error, 1) // Buffered to avoid blocking if the reader isn't ready

	var obj T
	value := reflect.ValueOf(obj)

	for value.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("Type T is not a struct")
	}

	typ := value.Type()

	go func() {
		defer close(outCh)
		defer close(errCh)

		headerMap := make(map[string]int)
		reader := csv.NewReader(reader)

		// Skip header row if present
		if t.HasHeader {
			header, err := reader.Read()
			if err != nil {
				errCh <- fmt.Errorf("error reading header: %w", err)
				return
			}

			for i, key := range header {
				headerMap[key] = i
			}
		}

		for {
			row, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					break // End of file
				}
				errCh <- fmt.Errorf("error reading record: %w", err)
				return
			}

			var obj T
			if t.HasHeader {
				obj, err = t.unmarshalNameTag(typ, headerMap, row)
			} else {
				obj, err = t.unmarshalNumTag(typ, row)
			}

			if err != nil {
				errCh <- err
				continue
			}

			outCh <- obj
		}
	}()

	return outCh, errCh, nil
}

func (t *Reader[T]) unmarshalNumTag(typ reflect.Type, row []string) (T, error) {
	var obj T
	value := reflect.New(typ).Elem()

	// Iterate over fields in the struct
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		colTag := field.Tag.Get(CSVTag)

		colIndex, err := strconv.Atoi(colTag)
		if err != nil {
			return obj, fmt.Errorf("csvTag should be an int %v", err)
		}

		if colIndex >= len(row) {
			continue
		}

		if err := t.setField(&value, field, i, row[colIndex]); err != nil {
			return obj, err
		}
	}

	return value.Interface().(T), nil
}

func (t *Reader[T]) unmarshalNameTag(typ reflect.Type, headerMap map[string]int, row []string) (T, error) {
	var obj T
	value := reflect.New(typ).Elem()

	// Iterate over fields in the struct
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		colTag := field.Tag.Get(CSVTag)

		colIndex, found := headerMap[colTag]
		if !found {
			continue
		}

		if colIndex >= len(row) {
			continue
		}

		if err := t.setField(&value, field, i, row[colIndex]); err != nil {
			return obj, err
		}
	}

	return value.Interface().(T), nil
}

func (t *Reader[T]) setField(v *reflect.Value, field reflect.StructField, i int, str string) error {
	switch field.Type.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(str)
		if err != nil {
			return errors.Wrapf(err, "invalid string float %v", v)
		}
		v.Field(i).SetBool(b)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return errors.Wrapf(err, "invalid string float %v", v)
		}
		v.Field(i).SetFloat(f)

	case reflect.String:
		fmt.Println("string", str)
		v.Field(i).SetString(str)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "invalid string int64 %v", v)
		}
		v.Field(i).SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		v.Field(i).SetUint(uintValue)

	case reflect.Struct:
		switch field.Type.Name() {
		case "Time":
			format := t.DateFormat

			date, err := time.Parse(format, str)
			if err != nil {
				return errors.Wrapf(err, "invalid string date %v", v)
			}

			v.Field(i).Set(reflect.ValueOf(date))
		}

	default:
		fmt.Println("None", field, field.Type.Name())
	}

	return nil
}

/*
// findColumnIndex finds the index of a column in the header by name, returns -1 if not found
func findColumnIndex(header []string, colName string) int {
	for i, name := range header {
		if name == colName {
			return i
		}
	}
	return -1
}

// setFieldValue sets the value of a struct field based on the CSV data, handling type conversions
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}
	return nil
}
*/

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

/*
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
*/
