package file_storage

import (
	"fmt"
	"os"
	"time"
)

type FileStorage struct {
	basePath string
}

func New(basePath string) (*FileStorage, error) {
	const op = "file_storage.fs.New"

	err := os.MkdirAll(basePath, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &FileStorage{
		basePath: basePath,
	}, nil
}

func (f *FileStorage) SaveImage(image []byte) (string, error) {
	const op = "file_storage.fs.SaveImage"

	filename := time.Now().Unix()
	err := os.WriteFile(fmt.Sprintf("%s/%d.png", f.basePath, filename), image, 0600)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return fmt.Sprintf("%d.png", filename), nil
}
