package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"crypto/sha256"
	"encoding/hex"
)

// Storage defines the interface for file storage backends
type Storage interface {
	PresignPut(ctx context.Context, objectName, contentType string, expiresIn time.Duration) (string, error)
	PresignGet(ctx context.Context, objectName string, expiresIn time.Duration) (string, error)
	Put(ctx context.Context, objectName string, reader io.Reader) error
	Get(ctx context.Context, objectName string) (io.ReadCloser, error)
	Delete(ctx context.Context, objectName string) error
}

// LocalStorage implements Storage using local filesystem
type LocalStorage struct {
	baseDir string
	baseURL string
}

// NewLocalStorage creates a new local filesystem storage backend
func NewLocalStorage(baseDir, baseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}
	return &LocalStorage{
		baseDir: baseDir,
		baseURL: baseURL,
	}, nil
}

func (s *LocalStorage) PresignPut(ctx context.Context, objectName, contentType string, expiresIn time.Duration) (string, error) {
	// For local storage, we return a URL that the client can PUT to
	// In a real implementation, you might want to generate a temporary token
	return fmt.Sprintf("%s/files/%s", s.baseURL, objectName), nil
}

func (s *LocalStorage) PresignGet(ctx context.Context, objectName string, expiresIn time.Duration) (string, error) {
	return fmt.Sprintf("%s/files/%s", s.baseURL, objectName), nil
}

func (s *LocalStorage) Put(ctx context.Context, objectName string, reader io.Reader) error {
	fullPath := filepath.Join(s.baseDir, objectName)
	
	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (s *LocalStorage) Get(ctx context.Context, objectName string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.baseDir, objectName)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (s *LocalStorage) Delete(ctx context.Context, objectName string) error {
	fullPath := filepath.Join(s.baseDir, objectName)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// CalculateSHA256 calculates SHA256 hash of file content
func CalculateSHA256(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

