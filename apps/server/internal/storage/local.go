// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PutObjectRequest struct {
	Bucket      string
	ObjectKey   string
	ContentType string
	Reader      io.Reader
}

type ObjectDescriptor struct {
	StorageBackend string    `json:"storageBackend"`
	Bucket         string    `json:"bucket"`
	ObjectKey      string    `json:"objectKey"`
	ContentType    string    `json:"contentType"`
	SizeBytes      int64     `json:"sizeBytes"`
	CreatedAt      time.Time `json:"createdAt"`
}

type ObjectStorage interface {
	Put(ctx context.Context, req PutObjectRequest) (ObjectDescriptor, error)
}

type LocalStorage struct {
	rootDir string
}

func NewLocalStorage(rootDir string) *LocalStorage {
	return &LocalStorage{rootDir: rootDir}
}

func (s *LocalStorage) Put(_ context.Context, req PutObjectRequest) (ObjectDescriptor, error) {
	bucket := strings.TrimSpace(req.Bucket)
	if bucket == "" {
		return ObjectDescriptor{}, fmt.Errorf("bucket is required")
	}
	if strings.TrimSpace(req.ObjectKey) == "" {
		return ObjectDescriptor{}, fmt.Errorf("objectKey is required")
	}
	if req.Reader == nil {
		return ObjectDescriptor{}, fmt.Errorf("reader is required")
	}

	relativePath := filepath.Clean(req.ObjectKey)
	if filepath.IsAbs(relativePath) || strings.HasPrefix(relativePath, "..") {
		return ObjectDescriptor{}, fmt.Errorf("objectKey must stay within storage root")
	}

	if filepath.IsAbs(bucket) || strings.Contains(bucket, "..") || strings.ContainsRune(bucket, filepath.Separator) {
		return ObjectDescriptor{}, fmt.Errorf("bucket must be a simple logical name")
	}

	targetPath := filepath.Join(s.rootDir, bucket, relativePath)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return ObjectDescriptor{}, fmt.Errorf("create object directory: %w", err)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return ObjectDescriptor{}, fmt.Errorf("create object file: %w", err)
	}
	defer file.Close()

	size, err := io.Copy(file, req.Reader)
	if err != nil {
		return ObjectDescriptor{}, fmt.Errorf("write object file: %w", err)
	}

	return ObjectDescriptor{
		StorageBackend: "local",
		Bucket:         bucket,
		ObjectKey:      relativePath,
		ContentType:    req.ContentType,
		SizeBytes:      size,
		CreatedAt:      time.Now().UTC(),
	}, nil
}
