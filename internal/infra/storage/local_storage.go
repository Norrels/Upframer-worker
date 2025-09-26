package storage

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"upframer-worker/internal/domain/ports"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
	}
}

func (ls *LocalStorage) Store(filePath, fileName string) (*ports.StorageResult, error) {
	err := os.MkdirAll(ls.basePath, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating base directory: %v", err)
	}

	destPath := filepath.Join(ls.basePath, fileName)

	src, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening source file: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("error creating destination file: %v", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("error copying file: %v", err)
	}

	return &ports.StorageResult{
		Path: destPath,
		URL:  fmt.Sprintf("file://%s", destPath),
	}, nil
}

func (ls *LocalStorage) StoreZip(sourceDir, zipFileName string) (*ports.StorageResult, error) {
	err := os.MkdirAll(ls.basePath, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating base directory: %v", err)
	}

	zipPath := filepath.Join(ls.basePath, zipFileName)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return nil, fmt.Errorf("error creating zip file: %v", err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		writer, err := archive.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("error creating zip: %v", err)
	}

	return &ports.StorageResult{
		Path: zipPath,
		URL:  fmt.Sprintf("file://%s", zipPath),
	}, nil
}