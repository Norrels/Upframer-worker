package ports

type StorageResult struct {
	Path string
	URL  string
}

type Storage interface {
	StoreZip(sourceDir, zipFileName string) (*StorageResult, error)
	Download(path, localPath string) error
}
