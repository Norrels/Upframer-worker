package ports

type StorageResult struct {
	Path string
	URL  string
}

type Storage interface {
	Store(filePath, fileName string) (*StorageResult, error)
	StoreZip(sourceDir, zipFileName string) (*StorageResult, error)
}
