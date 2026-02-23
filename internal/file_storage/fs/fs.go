package file_storage

import (
	"fmt"
	"os"
)

type FileStorage struct {
	basePath   string
	serverAddr string
}

func New(basePath string, serverAddr string) (*FileStorage, error) {
	const op = "file_storage.fs.New"

	err := os.MkdirAll(basePath, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &FileStorage{
		basePath:   basePath,
		serverAddr: serverAddr,
	}, nil
}

func (f *FileStorage) SaveImage(id int64, image []byte) (string, error) {
	const op = "file_storage.fs.SaveImage"

	err := os.WriteFile(fmt.Sprintf("%s/%d.png", f.basePath, id), image, 0600)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return fmt.Sprintf("http://%s/%d.png", f.serverAddr, id), nil
}

func (f *FileStorage) DeleteImage(id int64) error {
	const op = "file_storage.fs.DeleteImage"

	err := os.Remove(fmt.Sprintf("%s/%d.png", f.basePath, id))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
