package gdrive_helpers

import (
	"os"
	"path/filepath"
)

func Glob(dir string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	return files, err
}
