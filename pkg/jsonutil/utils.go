package jsonutil

import (
	"encoding/json"
	"io"
	"os"
)

func ToChannel[T any](filename string) (<-chan T, <-chan error, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	out := make(chan T)
	errCh := make(chan error)
	go func() {
		defer close(out)

		for i := 1; ; i++ {
			var obj T
			err := decoder.Decode(&obj)

			if err == io.EOF {
				break
			}

			if err != nil {
				errCh <- err
			}

			out <- obj
		}
	}()
	return out, nil, nil
}

func ToList[T any](filename string) ([]T, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	list := make([]T, 0, 100)
	for i := 1; ; i++ {
		var obj T
		err := decoder.Decode(&obj)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		list = append(list, obj)
	}

	return list, nil
}
