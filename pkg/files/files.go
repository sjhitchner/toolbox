package files

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func FindFiles(directoryPath string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				out <- path
			}
			return nil
		})

		if err != nil {
			panic(fmt.Errorf("Error walking the directory: %v", err))
		}
	}()

	return out
}

func IsS3(path string) (string, string, bool) {
	if strings.HasPrefix(path, "s3://") {
		u, err := url.Parse(path)
		if err != nil {
			return "", "", false
		}
		return u.Host, u.Path[1:], true
	}
	return "", "", false
}
