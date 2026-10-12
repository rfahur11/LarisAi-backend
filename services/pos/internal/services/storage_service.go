package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type StorageService interface {
	UploadProductImage(ctx context.Context, originalName string, reader io.Reader) (string, error)
}

type storageService struct {
	client        *azblob.Client
	containerName string
	accountName   string
}

func NewStorageService() StorageService {
	connStr := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
	if accountName == "" {
		accountName = "larisaimedia"
	}
	containerName := os.Getenv("AZURE_STORAGE_CONTAINER")
	if containerName == "" {
		containerName = "products-media"
	}

	if connStr != "" {
		client, err := azblob.NewClientFromConnectionString(connStr, nil)
		if err == nil {
			return &storageService{
				client:        client,
				containerName: containerName,
				accountName:   accountName,
			}
		}
	}

	return &storageService{
		containerName: containerName,
		accountName:   accountName,
	}
}

func (s *storageService) UploadProductImage(ctx context.Context, originalName string, reader io.Reader) (string, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	if ext == "" {
		ext = ".jpg"
	}
	// Sanitize filename
	cleanBase := strings.ReplaceAll(filepath.Base(originalName), ext, "")
	cleanBase = strings.ReplaceAll(cleanBase, " ", "_")
	blobName := fmt.Sprintf("prod_%d_%s%s", time.Now().UnixNano(), cleanBase, ext)

	// 1. If Azure Blob client is active
	if s.client != nil {
		data, err := io.ReadAll(reader)
		if err != nil {
			return "", err
		}

		_, err = s.client.UploadBuffer(ctx, s.containerName, blobName, data, &azblob.UploadBufferOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to upload to azure blob: %w", err)
		}

		publicURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.accountName, s.containerName, blobName)
		return publicURL, nil
	}

	// 2. Fallback to local storage
	uploadDir := "./uploads"
	_ = os.MkdirAll(uploadDir, 0755)
	destPath := filepath.Join(uploadDir, blobName)
	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, reader); err != nil {
		return "", err
	}

	return fmt.Sprintf("/api/v1/images/%s", blobName), nil
}
