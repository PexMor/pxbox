package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"pxbox/internal/storage"
)

func (d Dependencies) signFile(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	contentType := r.URL.Query().Get("contentType")
	requestID := r.URL.Query().Get("requestId")
	fileSizeStr := r.URL.Query().Get("size") // File size in bytes (optional, for validation)

	if name == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "name parameter required", d.Log)
		return
	}

	// If requestId is provided, validate against request's file policy
	if requestID != "" {
		req, err := d.DB.Queries.GetRequestByID(r.Context(), requestID)
		if err != nil {
			WriteError(w, http.StatusNotFound, "request_not_found", "Request not found", d.Log)
			return
		}

		// Parse and validate file policy
		if req.FilesPolicy != nil {
			policy, err := storage.ParseFilePolicy(req.FilesPolicy)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "invalid_policy", "Invalid file policy", d.Log)
				return
			}

			// Validate file size if provided
			if fileSizeStr != "" && policy != nil {
				fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
				if err != nil {
					WriteError(w, http.StatusBadRequest, "invalid_size", "Invalid file size parameter", d.Log)
					return
				}

				if err := policy.ValidateFile(name, contentType, fileSize); err != nil {
					WriteError(w, http.StatusBadRequest, "policy_violation", err.Error(), d.Log)
					return
				}
			} else if policy != nil {
				// If size not provided but policy exists, validate MIME type and extension only
				if err := policy.ValidateFile(name, contentType, 0); err != nil {
					// Only fail if it's a MIME type or extension error (not size)
					if !strings.Contains(err.Error(), "exceeds maximum") {
						WriteError(w, http.StatusBadRequest, "policy_violation", err.Error(), d.Log)
						return
					}
				}
			}
		}
	}

	// Initialize storage (local filesystem for now)
	baseDir := os.Getenv("STORAGE_BASE_DIR")
	if baseDir == "" {
		baseDir = "./storage"
	}
	baseURL := os.Getenv("STORAGE_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	stor, err := storage.NewLocalStorage(baseDir, baseURL)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "storage_init_failed", "Storage initialization failed", d.Log)
		return
	}

	putURL, err := stor.PresignPut(r.Context(), name, contentType, 15*time.Minute)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "url_generation_failed", "Failed to generate presigned URL", d.Log)
		return
	}

	getURL, err := stor.PresignGet(r.Context(), name, 24*time.Hour)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "url_generation_failed", "Failed to generate presigned URL", d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"putUrl": putURL,
		"getUrl": getURL,
	})
}

