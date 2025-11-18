package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// StreamEvent represents an event from streams
type StreamEvent struct {
	Channel   string
	Sequence  int64
	Event     map[string]interface{}
	Timestamp time.Time
}

// StreamsProvider interface for event replay
type StreamsProvider interface {
	GetLastSequence(channel, connectionID string) (int64, error)
	AcknowledgeSequence(channel, connectionID string, sequence int64) error
	ReplayEvents(channel string, sinceSeq int64, limit int64) ([]StreamEvent, error)
}

// Hub manages WebSocket connections and channel subscriptions
type Hub struct {
	mu         sync.RWMutex
	conns      map[*Conn]bool
	subs       map[string]map[*Conn]bool // channel -> connections
	publish    chan Event
	log        *zap.Logger
	cmdHandler *CommandHandler
	ctx        context.Context
	streams    StreamsProvider // For sequence numbers and replay
}

// Conn represents a WebSocket connection
type Conn struct {
	ws     *websocket.Conn
	send   chan []byte
	hub    *Hub
	userID string
	subs   map[string]bool // subscribed channels
	ctx    context.Context
}

// Event represents a message to be published
type Event struct {
	Channel string
	Message map[string]interface{}
}

// NewHub creates a new WebSocket hub
func NewHub(log *zap.Logger) *Hub {
	return &Hub{
		conns:   make(map[*Conn]bool),
		subs:    make(map[string]map[*Conn]bool),
		publish: make(chan Event, 256),
		log:     log,
		ctx:     context.Background(),
	}
}

// SetCommandHandler sets the command handler for processing WebSocket commands
func (h *Hub) SetCommandHandler(handler *CommandHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cmdHandler = handler
}

// SetStreamsProvider sets the streams provider for event replay
func (h *Hub) SetStreamsProvider(provider StreamsProvider) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.streams = provider
}

// Run starts the hub's event loop
func (h *Hub) Run() {
	for event := range h.publish {
		h.mu.RLock()
		conns := h.subs[event.Channel]
		h.mu.RUnlock()

		if conns != nil {
			msg, _ := json.Marshal(event.Message)
			for conn := range conns {
				select {
				case conn.send <- msg:
				default:
					close(conn.send)
					h.unregister(conn)
				}
			}
		}
	}
}

// Register adds a new connection to the hub
func (h *Hub) Register(conn *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[conn] = true
}

// Unregister removes a connection from the hub
func (h *Hub) unregister(conn *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.conns[conn]; ok {
		delete(h.conns, conn)
		close(conn.send)
		for channel := range conn.subs {
			if subs := h.subs[channel]; subs != nil {
				delete(subs, conn)
				if len(subs) == 0 {
					delete(h.subs, channel)
				}
			}
		}
	}
}

// Subscribe adds a connection to a channel
func (h *Hub) Subscribe(conn *Conn, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs[channel] == nil {
		h.subs[channel] = make(map[*Conn]bool)
	}
	h.subs[channel][conn] = true
	conn.subs[channel] = true
}

// Unsubscribe removes a connection from a channel
func (h *Hub) Unsubscribe(conn *Conn, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if subs := h.subs[channel]; subs != nil {
		delete(subs, conn)
		if len(subs) == 0 {
			delete(h.subs, channel)
		}
	}
	delete(conn.subs, channel)
}

// Publish sends an event to all subscribers of a channel
func (h *Hub) Publish(channel string, message map[string]interface{}) {
	select {
	case h.publish <- Event{Channel: channel, Message: message}:
	default:
		h.log.Warn("Hub publish channel full, dropping event", zap.String("channel", channel))
	}
}

// NewConn creates a new connection
func NewConn(ws *websocket.Conn, hub *Hub, userID string) *Conn {
	return &Conn{
		ws:     ws,
		send:   make(chan []byte, 256),
		hub:    hub,
		userID: userID,
		subs:   make(map[string]bool),
		ctx:    hub.ctx,
	}
}

// ReadPump handles reading from the WebSocket connection
func (c *Conn) ReadPump() {
	defer func() {
		c.hub.unregister(c)
		c.ws.Close()
	}()

	c.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.log.Error("WebSocket error", zap.Error(err))
			}
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			c.hub.log.Warn("Failed to parse message", zap.Error(err))
			continue
		}

		c.handleMessage(msg)
	}
}

// WritePump handles writing to the WebSocket connection
func (c *Conn) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Conn) handleMessage(msg map[string]interface{}) {
	msgType, _ := msg["type"].(string)
	
	switch msgType {
	case "subscribe":
		channel, _ := msg["channel"].(string)
		if channel != "" {
			c.hub.Subscribe(c, channel)
			c.sendAck("subscribed", channel)
		}
	case "unsubscribe":
		channel, _ := msg["channel"].(string)
		if channel != "" {
			c.hub.Unsubscribe(c, channel)
			c.sendAck("unsubscribed", channel)
		}
	case "ack":
		// Handle acknowledgment
		channel, _ := msg["channel"].(string)
		seq, _ := msg["seq"].(float64)
		if channel != "" && seq > 0 {
			c.hub.Acknowledge(c, channel, int64(seq))
		}
	case "resume":
		// Handle resume request
		channel, _ := msg["channel"].(string)
		since, _ := msg["since"].(float64)
		if channel != "" && since >= 0 {
			c.hub.Resume(c, channel, int64(since))
		}
	case "cmd":
		if c.hub.cmdHandler != nil {
			c.hub.cmdHandler.HandleCommand(c.ctx, c, msg)
		} else {
			c.hub.log.Warn("Command handler not set")
		}
	case "ping":
		c.sendAck("pong", "")
	default:
		c.hub.log.Warn("Unknown message type", zap.String("type", msgType))
	}
}

func (c *Conn) sendAck(msgType, channel string) {
	ack := map[string]interface{}{
		"type": "ack",
		"ack":  msgType,
	}
	if channel != "" {
		ack["channel"] = channel
	}
	msg, _ := json.Marshal(ack)
	select {
	case c.send <- msg:
	default:
	}
}

// Acknowledge records an acknowledgment for a sequence number
func (h *Hub) Acknowledge(conn *Conn, channel string, sequence int64) {
	if h.streams != nil {
		connectionID := conn.userID // Use userID as connection identifier
		if err := h.streams.AcknowledgeSequence(channel, connectionID, sequence); err != nil {
			h.log.Warn("Failed to acknowledge sequence",
				zap.String("channel", channel),
				zap.Int64("sequence", sequence),
				zap.Error(err),
			)
		}
	}
}

// Resume replays events from a given sequence number
func (h *Hub) Resume(conn *Conn, channel string, sinceSeq int64) {
	if h.streams == nil {
		h.log.Warn("Streams provider not set, cannot resume")
		return
	}
	
	events, err := h.streams.ReplayEvents(channel, sinceSeq, 100) // Limit to 100 events
	if err != nil {
		h.log.Error("Failed to replay events",
			zap.String("channel", channel),
			zap.Int64("since", sinceSeq),
			zap.Error(err),
		)
		return
	}
	
	// Send replayed events to connection
	for _, event := range events {
		msg := map[string]interface{}{
			"type":    "event",
			"channel": event.Channel,
			"seq":     event.Sequence,
			"data":    event.Event,
		}
		msgBytes, _ := json.Marshal(msg)
		select {
		case conn.send <- msgBytes:
		default:
			h.log.Warn("Failed to send replayed event, connection buffer full")
			return
		}
	}
	
	h.log.Info("Resumed events",
		zap.String("channel", channel),
		zap.String("connection", conn.userID),
		zap.Int64("since", sinceSeq),
		zap.Int("count", len(events)),
	)
}

