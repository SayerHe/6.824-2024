package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func FindFilesStartWith(root, prefix string) ([]string, error) {
	files := make([]string, 0)
	// path 为文件 path
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(info.Name(), prefix) {
			files = append(files, "../"+info.Name())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return files, nil
}

func Int32Ptr(number int) *int { return &number }
