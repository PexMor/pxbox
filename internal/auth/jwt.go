package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userID"
const entityIDKey contextKey = "entityID"

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey string
}

// NewJWTConfig creates a new JWT config
func NewJWTConfig(secretKey string) *JWTConfig {
	if secretKey == "" {
		secretKey = "default-secret-key-change-in-production" // Default for development
	}
	return &JWTConfig{SecretKey: secretKey}
}

// Middleware creates a JWT authentication middleware
func (c *JWTConfig) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header or X-Entity-ID header (for development)
		// In production, use JWT token from Authorization header
		entityID := r.Header.Get("X-Entity-ID")
		if entityID != "" {
			// Development mode: allow X-Entity-ID header
			ctx := context.WithValue(r.Context(), entityIDKey, entityID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Allow anonymous access for now (can be made stricter)
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(c.SecretKey), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userID, _ := claims["sub"].(string)
			entityID, _ := claims["entity_id"].(string)
			
			ctx := r.Context()
			if userID != "" {
				ctx = context.WithValue(ctx, userIDKey, userID)
			}
			if entityID != "" {
				ctx = context.WithValue(ctx, entityIDKey, entityID)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
	})
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetEntityID extracts entity ID from context
func GetEntityID(ctx context.Context) string {
	if entityID, ok := ctx.Value(entityIDKey).(string); ok {
		return entityID
	}
	return ""
}

