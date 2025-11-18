package storage

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"
)

// FilePolicy represents file upload policy constraints
type FilePolicy struct {
	MaxTotalMB *float64  `json:"maxTotalMB,omitempty"`
	MaxFileMB  *float64  `json:"maxFileMB,omitempty"`
	MimeTypes  []string  `json:"mime,omitempty"`
	Extensions []string  `json:"extensions,omitempty"`
}

// ParseFilePolicy parses a map[string]interface{} into FilePolicy
func ParseFilePolicy(policy map[string]interface{}) (*FilePolicy, error) {
	if policy == nil {
		return nil, nil
	}

	fp := &FilePolicy{}

	// Parse maxTotalMB
	if val, ok := policy["maxTotalMB"].(float64); ok {
		fp.MaxTotalMB = &val
	}

	// Parse maxFileMB
	if val, ok := policy["maxFileMB"].(float64); ok {
		fp.MaxFileMB = &val
	}

	// Parse mime types
	if mimeVal, ok := policy["mime"].([]interface{}); ok {
		fp.MimeTypes = make([]string, 0, len(mimeVal))
		for _, m := range mimeVal {
			if mStr, ok := m.(string); ok {
				fp.MimeTypes = append(fp.MimeTypes, mStr)
			}
		}
	}

	// Parse extensions
	if extVal, ok := policy["extensions"].([]interface{}); ok {
		fp.Extensions = make([]string, 0, len(extVal))
		for _, e := range extVal {
			if eStr, ok := e.(string); ok {
				// Normalize extensions (remove leading dot if present)
				ext := strings.TrimPrefix(eStr, ".")
				fp.Extensions = append(fp.Extensions, strings.ToLower(ext))
			}
		}
	}

	return fp, nil
}

// ValidateFile validates a file against the policy
func (fp *FilePolicy) ValidateFile(fileName, contentType string, fileSizeBytes int64) error {
	if fp == nil {
		return nil // No policy means no restrictions
	}

	// Validate file size
	if fp.MaxFileMB != nil {
		maxBytes := int64(*fp.MaxFileMB * 1024 * 1024)
		if fileSizeBytes > maxBytes {
			return fmt.Errorf("file size %d bytes exceeds maximum %d bytes (%.2f MB)", 
				fileSizeBytes, maxBytes, *fp.MaxFileMB)
		}
	}

	// Validate MIME type
	if len(fp.MimeTypes) > 0 {
		if !fp.matchesMimeType(contentType) {
			return fmt.Errorf("content type %s is not allowed. Allowed types: %v", 
				contentType, fp.MimeTypes)
		}
	}

	// Validate extension
	if len(fp.Extensions) > 0 {
		if !fp.matchesExtension(fileName) {
			return fmt.Errorf("file extension is not allowed. Allowed extensions: %v", 
				fp.Extensions)
		}
	}

	return nil
}

// matchesMimeType checks if contentType matches any of the allowed MIME type patterns
func (fp *FilePolicy) matchesMimeType(contentType string) bool {
	// Parse the content type (handle parameters like "image/png; charset=utf-8")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		// If parsing fails, use the original string
		mediaType = contentType
	}

	for _, allowed := range fp.MimeTypes {
		// Support wildcard patterns like "image/*"
		if strings.HasSuffix(allowed, "/*") {
			prefix := strings.TrimSuffix(allowed, "/*")
			if strings.HasPrefix(mediaType, prefix+"/") {
				return true
			}
		} else if mediaType == allowed {
			return true
		}
	}
	return false
}

// matchesExtension checks if fileName has an allowed extension
func (fp *FilePolicy) matchesExtension(fileName string) bool {
	ext := strings.ToLower(strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), "."))
	if ext == "" {
		return false
	}

	for _, allowed := range fp.Extensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

