package file_storage

import (
	"fmt"
	"os"

	"github.com/google/uuid"
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

func (f *FileStorage) SaveImage(id uuid.UUID, image []byte) (string, error) {
	const op = "file_storage.fs.SaveImage"

	err := os.WriteFile(fmt.Sprintf("%s/%s.png", f.basePath, id.String()), image, 0600)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return fmt.Sprintf("http://%s/%s.png", f.serverAddr, id.String()), nil
}

func (f *FileStorage) DeleteImage(id uuid.UUID) error {
	const op = "file_storage.fs.DeleteImage"

	err := os.Remove(fmt.Sprintf("%s/%s.png", f.basePath, id.String()))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
