package pubsub

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Bus struct {
	rdb     *redis.Client
	log     *zap.Logger
	ctx     context.Context
	wsHub   WSHub
	streams *Streams
}

type WSHub interface {
	Publish(channel string, message map[string]interface{})
}

func New(rdb *redis.Client, log *zap.Logger) *Bus {
	return &Bus{
		rdb:     rdb,
		log:     log,
		ctx:     context.Background(),
		streams: NewStreams(rdb, log),
	}
}

// SetWSHub sets the WebSocket hub for event broadcasting
func (b *Bus) SetWSHub(hub WSHub) {
	b.wsHub = hub
}

// GetStreams returns the streams provider
func (b *Bus) GetStreams() *Streams {
	return b.streams
}

// PublishEntity publishes an event to an entity's channel
func (b *Bus) PublishEntity(entityID string, event map[string]interface{}) error {
	channel := "entity:" + entityID
	return b.Publish(channel, event)
}

// PublishRequest publishes an event to a request's channel
func (b *Bus) PublishRequest(requestID string, event map[string]interface{}) error {
	channel := "request:" + requestID
	return b.Publish(channel, event)
}

// PublishRequestor publishes an event to a requestor's channel
func (b *Bus) PublishRequestor(clientID string, event map[string]interface{}) error {
	channel := "requestor:" + clientID
	return b.Publish(channel, event)
}

// Publish publishes an event to a channel
func (b *Bus) Publish(channel string, event map[string]interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Publish to Redis pub/sub
	err = b.rdb.Publish(b.ctx, channel, data).Err()
	if err != nil {
		b.log.Error("Failed to publish event", zap.String("channel", channel), zap.Error(err))
		return err
	}

	// Also publish to Redis Streams for replay
	seq, err := b.streams.PublishEvent(channel, event)
	if err != nil {
		b.log.Warn("Failed to publish to stream", zap.String("channel", channel), zap.Error(err))
		// Continue even if stream publish fails
	}

	// Add sequence number to event for WebSocket
	eventWithSeq := make(map[string]interface{})
	for k, v := range event {
		eventWithSeq[k] = v
	}
	eventWithSeq["seq"] = seq

	// Broadcast to WebSocket hub if available
	if b.wsHub != nil {
		b.wsHub.Publish(channel, eventWithSeq)
	}

	b.log.Debug("Published event", zap.String("channel", channel), zap.Int64("seq", seq), zap.String("event", string(data)))
	return nil
}

