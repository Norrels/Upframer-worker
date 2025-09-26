package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"upframer-worker/internal/domain/ports"
	"upframer-worker/internal/infra/util"
)

type LocalStorage struct {
	basePath   string
	zipAdapter *util.ZipAdapter
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{
		basePath:   basePath,
		zipAdapter: util.NewZipAdapter(),
	}
}

func (ls *LocalStorage) StoreZip(sourceDir, zipFileName string) (*ports.StorageResult, error) {
	err := os.MkdirAll(ls.basePath, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating base directory: %v", err)
	}

	zipPath := filepath.Join(ls.basePath, zipFileName)

	err = ls.zipAdapter.CreateZipFile(sourceDir, zipPath)
	if err != nil {
		return nil, fmt.Errorf("error creating zip: %v", err)
	}

	return &ports.StorageResult{
		Path: zipPath,
		URL:  fmt.Sprintf("file://%s", zipPath),
	}, nil
}

func (ls *LocalStorage) Download(s3Key, localPath string) error {
	return fmt.Errorf("Download is not supported for LocalStorage")
}