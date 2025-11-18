package api

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// Error writes a standardized error response
func WriteError(w http.ResponseWriter, code int, errCode, message string, log *zap.Logger) {
	log.Error("API error", zap.String("code", errCode), zap.String("message", message))
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	resp := ErrorResponse{
		Error:   errCode,
		Message: message,
	}
	if errCode != "" {
		resp.Code = errCode
	}
	
	json.NewEncoder(w).Encode(resp)
}

// RequestLogger logs HTTP requests and responses
func RequestLogger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip wrapping for WebSocket upgrades - they need direct access to ResponseWriter
			if r.Header.Get("Upgrade") == "websocket" {
				next.ServeHTTP(w, r)
				return
			}
			
			start := time.Now()
			
			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			next.ServeHTTP(wrapped, r)
			
			duration := time.Since(start)
			
			log.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", duration),
				zap.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

