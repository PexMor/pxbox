package service

import (
	"context"
	"testing"
	"time"

	"pxbox/internal/db"
	"pxbox/internal/model"
	"pxbox/internal/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEventBus implements EventBus for testing
type MockEventBus struct {
	events []map[string]interface{}
}

func (m *MockEventBus) PublishEntity(entityID string, event map[string]interface{}) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventBus) PublishRequest(requestID string, event map[string]interface{}) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventBus) PublishRequestor(clientID string, event map[string]interface{}) error {
	m.events = append(m.events, event)
	return nil
}

func TestRequestService_CreateRequest(t *testing.T) {
	t.Skip("Requires test database setup")
}

func TestRequestService_GetRequest(t *testing.T) {
	t.Skip("Requires test database setup")
}

func TestRequestService_ClaimRequest(t *testing.T) {
	t.Skip("Requires test database setup")
}

func TestRequestService_PostResponse(t *testing.T) {
	t.Skip("Requires test database setup")
}

func TestRequestService_CancelRequest(t *testing.T) {
	t.Skip("Requires test database setup")
}

