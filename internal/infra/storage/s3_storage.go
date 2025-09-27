package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"upframer-worker/internal/domain/ports"
	"upframer-worker/internal/infra/util"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	bucket     string
	client     *s3.Client
	zipAdapter *util.ZipAdapter
}

func NewS3Storage(bucket, region, accessKey, secretKey, sessionToken string) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Storage{
		bucket:     bucket,
		client:     client,
		zipAdapter: util.NewZipAdapter(),
	}, nil
}

func (s *S3Storage) StoreZip(sourceDir, zipFileName string) (*ports.StorageResult, error) {
	tempDir, err := os.MkdirTemp("", "upframer-secure-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create secure temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chmod(tempDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to set secure permissions: %v", err)
	}

	tempZipPath := filepath.Join(tempDir, zipFileName)

	err = s.zipAdapter.CreateZipFile(sourceDir, tempZipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create zip file: %v", err)
	}

	uploader := manager.NewUploader(s.client)

	file, err := os.Open(tempZipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %v", err)
	}
	defer file.Close()

	s3Key := fmt.Sprintf("results/%s", zipFileName)

	result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
		Body:   file,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload zip to S3: %v", err)
	}

	return &ports.StorageResult{
		Path: s3Key,
		URL:  result.Location,
	}, nil
}

func (s *S3Storage) Download(s3Key, localPath string) error {
	downloader := manager.NewDownloader(s.client)

	err := os.MkdirAll(filepath.Dir(localPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer file.Close()

	_, err = downloader.Download(context.TODO(), file, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to download file from S3: %v", err)
	}

	return nil
}
