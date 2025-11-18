package api

import (
	"net/http"
	"os"
	"strings"

	"pxbox/internal/ws"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking
		return true
	},
}

func (d Dependencies) wsHandler(w http.ResponseWriter, r *http.Request) {
	d.Log.Info("WebSocket connection attempt", 
		zap.String("remote", r.RemoteAddr),
		zap.String("path", r.URL.Path),
		zap.String("upgrade", r.Header.Get("Upgrade")),
	)

	// Check Hub before upgrading
	if d.Hub == nil {
		d.Log.Error("WebSocket hub not initialized")
		http.Error(w, "WebSocket hub not initialized", http.StatusInternalServerError)
		return
	}

	// Extract user ID from JWT token or header
	userID := extractUserIDFromRequest(r, d.Log)
	if userID == "" {
		userID = "anonymous"
	}
	d.Log.Info("WebSocket user ID", zap.String("userID", userID))

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		d.Log.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	d.Log.Info("WebSocket connection upgraded successfully")

	wsConn := ws.NewConn(conn, d.Hub, userID)
	d.Hub.Register(wsConn)

	go wsConn.WritePump()
	go wsConn.ReadPump()
}

func extractUserIDFromRequest(r *http.Request, log *zap.Logger) string {
	// Try JWT subprotocol first
	if subprotocols := websocket.Subprotocols(r); len(subprotocols) > 0 {
		for _, subprotocol := range subprotocols {
			if subprotocol == "jwt" {
				// Extract token from query parameter or header
				tokenString := r.URL.Query().Get("token")
				if tokenString == "" {
					tokenString = r.Header.Get("Authorization")
					if strings.HasPrefix(tokenString, "Bearer ") {
						tokenString = strings.TrimPrefix(tokenString, "Bearer ")
					}
				}
				
				if tokenString != "" {
					jwtSecret := os.Getenv("JWT_SECRET")
					if jwtSecret == "" {
						jwtSecret = "default-secret-key-change-in-production"
					}
					
					token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
						return []byte(jwtSecret), nil
					})
					
					if err == nil && token.Valid {
						if claims, ok := token.Claims.(jwt.MapClaims); ok {
							if entityID, ok := claims["entity_id"].(string); ok && entityID != "" {
								return entityID
							}
							if userID, ok := claims["sub"].(string); ok && userID != "" {
								return userID
							}
						}
					}
				}
			}
		}
	}
	
	// Fallback to X-Entity-ID header (for development)
	if entityID := r.Header.Get("X-Entity-ID"); entityID != "" {
		return entityID
	}
	
	return ""
}

