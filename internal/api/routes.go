package api

import (
	"net/http"
	"os"

	"pxbox/internal/auth"
	"pxbox/internal/db"
	"pxbox/internal/pubsub"
	"pxbox/internal/service"
	"pxbox/internal/ws"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Dependencies struct {
	DB        *db.Pool
	Bus       *pubsub.Bus
	Hub       *ws.Hub
	Log       *zap.Logger
	JobClient service.JobClient
}

func Routes(d Dependencies) http.Handler {
	r := chi.NewRouter()
	
	// Add request logging middleware
	r.Use(RequestLogger(d.Log))
	
	// Add JWT authentication middleware (optional - allows anonymous access)
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtConfig := auth.NewJWTConfig(jwtSecret)
	r.Use(jwtConfig.Middleware)

	// Request endpoints
	r.Post("/requests", d.createRequest)
	r.Get("/requests/{id}", d.getRequest)
	r.Post("/requests/{id}/cancel", d.cancelRequest)
	r.Post("/requests/{id}/claim", d.claimRequest)
	r.Post("/requests/{id}/response", d.postResponse)
	r.Get("/requests/{id}/response", d.getResponse)

	// Entity endpoints
	r.Post("/entities", d.createEntity)
	r.Get("/entities/{id}", d.getEntity)
	r.Get("/entities/{id}/queue", d.entityQueue)

	// Flow endpoints
	r.Post("/flows", d.createFlow)
	r.Get("/flows/{id}", d.getFlow)
	r.Post("/flows/{id}/resume", d.resumeFlow)
	r.Post("/flows/{id}/cancel", d.cancelFlow)

	// Inquiry endpoints
	r.Get("/inquiries", d.listInquiries)
	r.Post("/inquiries/{id}/markRead", d.markRead)
	r.Post("/inquiries/{id}/snooze", d.snooze)
	r.Post("/inquiries/{id}/cancel", d.cancelInquiry)
	r.Delete("/inquiries/{id}", d.deleteInquiry)

	// File endpoints
	r.Post("/files/sign", d.signFile)

	// WebSocket endpoint
	r.Get("/ws", d.wsHandler)

	return r
}
