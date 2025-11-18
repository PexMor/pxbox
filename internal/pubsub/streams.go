package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// StreamEvent represents an event stored in Redis Streams
// Note: This matches ws.StreamEvent structure
type StreamEvent struct {
	Channel   string
	Sequence  int64
	Event     map[string]interface{}
	Timestamp time.Time
}

// Streams manages Redis Streams for event replay
type Streams struct {
	rdb  *redis.Client
	log  *zap.Logger
	ctx  context.Context
}

// NewStreams creates a new Streams manager
func NewStreams(rdb *redis.Client, log *zap.Logger) *Streams {
	return &Streams{
		rdb: rdb,
		log: log,
		ctx: context.Background(),
	}
}

// PublishEvent publishes an event to a Redis Stream with sequence number
func (s *Streams) PublishEvent(channel string, event map[string]interface{}) (int64, error) {
	streamKey := fmt.Sprintf("stream:%s", channel)
	
	// Get next sequence number
	seq, err := s.getNextSequence(channel)
	if err != nil {
		return 0, fmt.Errorf("failed to get sequence: %w", err)
	}
	
	// Add sequence to event
	eventWithSeq := make(map[string]interface{})
	for k, v := range event {
		eventWithSeq[k] = v
	}
	eventWithSeq["seq"] = seq
	eventWithSeq["channel"] = channel
	eventWithSeq["timestamp"] = time.Now().Format(time.RFC3339)
	
	// Marshal event data
	eventData, err := json.Marshal(eventWithSeq)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal event: %w", err)
	}
	
	// Add to stream
	args := redis.XAddArgs{
		Stream: streamKey,
		ID:     "*", // Auto-generate ID
		Values: map[string]interface{}{
			"data": string(eventData),
		},
	}
	
	id, err := s.rdb.XAdd(s.ctx, &args).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to add to stream: %w", err)
	}
	
	// Parse sequence from ID (format: timestamp-sequence)
	seqFromID, _ := parseStreamID(id)
	if seqFromID > 0 {
		seq = seqFromID
	}
	
	s.log.Debug("Published event to stream",
		zap.String("channel", channel),
		zap.Int64("sequence", seq),
		zap.String("stream_id", id),
	)
	
	return seq, nil
}

// GetNextSequence gets the next sequence number for a channel
func (s *Streams) getNextSequence(channel string) (int64, error) {
	seqKey := fmt.Sprintf("seq:%s", channel)
	
	// Increment and get sequence
	seq, err := s.rdb.Incr(s.ctx, seqKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment sequence: %w", err)
	}
	
	return seq, nil
}

// GetLastSequence gets the last acknowledged sequence for a channel and connection
func (s *Streams) GetLastSequence(channel, connectionID string) (int64, error) {
	ackKey := fmt.Sprintf("ack:%s:%s", channel, connectionID)
	
	seqStr, err := s.rdb.Get(s.ctx, ackKey).Result()
	if err == redis.Nil {
		return 0, nil // No acknowledgment yet
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get last sequence: %w", err)
	}
	
	seq, err := strconv.ParseInt(seqStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse sequence: %w", err)
	}
	
	return seq, nil
}

// AcknowledgeSequence records an acknowledgment for a sequence number
func (s *Streams) AcknowledgeSequence(channel, connectionID string, sequence int64) error {
	ackKey := fmt.Sprintf("ack:%s:%s", channel, connectionID)
	
	err := s.rdb.Set(s.ctx, ackKey, sequence, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to acknowledge sequence: %w", err)
	}
	
	s.log.Debug("Acknowledged sequence",
		zap.String("channel", channel),
		zap.String("connection", connectionID),
		zap.Int64("sequence", sequence),
	)
	
	return nil
}

// ReplayEvents replays events from a given sequence number
func (s *Streams) ReplayEvents(channel string, sinceSeq int64, limit int64) ([]StreamEvent, error) {
	streamKey := fmt.Sprintf("stream:%s", channel)
	
	// Convert sequence to stream ID (approximate)
	// Format: timestamp-sequence (milliseconds-sequence)
	startID := fmt.Sprintf("%d-%d", time.Now().Add(-24*time.Hour).UnixMilli(), sinceSeq)
	
	args := redis.XReadArgs{
		Streams: []string{streamKey, startID},
		Count:   limit,
	}
	
	streams, err := s.rdb.XRead(s.ctx, &args).Result()
	if err == redis.Nil {
		return []StreamEvent{}, nil // No events
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read stream: %w", err)
	}
	
	var events []StreamEvent
	for _, stream := range streams {
		for _, msg := range stream.Messages {
			data, ok := msg.Values["data"].(string)
			if !ok {
				continue
			}
			
			var eventData map[string]interface{}
			if err := json.Unmarshal([]byte(data), &eventData); err != nil {
				s.log.Warn("Failed to unmarshal event", zap.Error(err))
				continue
			}
			
			seq, _ := eventData["seq"].(float64)
			channelName, _ := eventData["channel"].(string)
			timestampStr, _ := eventData["timestamp"].(string)
			
			var timestamp time.Time
			if timestampStr != "" {
				timestamp, _ = time.Parse(time.RFC3339, timestampStr)
			}
			if timestamp.IsZero() {
				timestamp = time.Now()
			}
			
			// Remove metadata from event
			event := make(map[string]interface{})
			for k, v := range eventData {
				if k != "seq" && k != "channel" && k != "timestamp" {
					event[k] = v
				}
			}
			
			events = append(events, StreamEvent{
				Channel:   channelName,
				Sequence:  int64(seq),
				Event:     event,
				Timestamp: timestamp,
			})
		}
	}
	
	return events, nil
}

// parseStreamID parses a Redis Stream ID (format: timestamp-sequence)
func parseStreamID(id string) (int64, error) {
	parts := splitStreamID(id)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid stream ID format")
	}
	
	seq, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse sequence: %w", err)
	}
	
	return seq, nil
}

// splitStreamID splits a Redis Stream ID into parts
func splitStreamID(id string) []string {
	// Redis Stream IDs are in format: timestamp-sequence
	// They can also have "-0" suffix for auto-generated IDs
	var parts []string
	lastDash := -1
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == '-' {
			lastDash = i
			break
		}
	}
	
	if lastDash > 0 {
		parts = append(parts, id[:lastDash], id[lastDash+1:])
	} else {
		parts = append(parts, id)
	}
	
	return parts
}

