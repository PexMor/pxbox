package storage

import (
	"fmt"
)

// FileMetadata represents file metadata structure
type FileMetadata struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Size    int64  `json:"size"`
	MIME    string `json:"mime"`
	SHA256  string `json:"sha256,omitempty"`
}

// NormalizeFileMetadata normalizes file metadata from a map
func NormalizeFileMetadata(file map[string]interface{}) FileMetadata {
	meta := FileMetadata{}
	
	if name, ok := file["name"].(string); ok {
		meta.Name = name
	}
	if url, ok := file["url"].(string); ok {
		meta.URL = url
	}
	if size, ok := file["size"].(float64); ok {
		meta.Size = int64(size)
	} else if size, ok := file["size"].(int64); ok {
		meta.Size = size
	} else if size, ok := file["size"].(int); ok {
		meta.Size = int64(size)
	}
	if mime, ok := file["mime"].(string); ok {
		meta.MIME = mime
	} else if contentType, ok := file["contentType"].(string); ok {
		meta.MIME = contentType
	}
	if sha256, ok := file["sha256"].(string); ok {
		meta.SHA256 = sha256
	}
	
	return meta
}

// ValidateFileMetadata validates that file metadata has required fields
func ValidateFileMetadata(meta FileMetadata) error {
	if meta.Name == "" {
		return fmt.Errorf("file name is required")
	}
	if meta.URL == "" {
		return fmt.Errorf("file URL is required")
	}
	if meta.Size < 0 {
		return fmt.Errorf("file size must be non-negative")
	}
	return nil
}

// ToMap converts FileMetadata to a map for storage
func (m FileMetadata) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"name": m.Name,
		"url":  m.URL,
		"size": m.Size,
		"mime": m.MIME,
	}
	if m.SHA256 != "" {
		result["sha256"] = m.SHA256
	}
	return result
}

// NormalizeFiles normalizes a slice of file metadata maps
func NormalizeFiles(files []map[string]interface{}) ([]map[string]interface{}, error) {
	normalized := make([]map[string]interface{}, 0, len(files))
	for _, file := range files {
		meta := NormalizeFileMetadata(file)
		if err := ValidateFileMetadata(meta); err != nil {
			return nil, fmt.Errorf("invalid file metadata: %w", err)
		}
		normalized = append(normalized, meta.ToMap())
	}
	return normalized, nil
}

