package model

// Status represents request status
type Status string

const (
	StatusPending  Status = "PENDING"
	StatusClaimed  Status = "CLAIMED"
	StatusAnswered Status = "ANSWERED"
	StatusCancelled Status = "CANCELLED"
	StatusExpired  Status = "EXPIRED"
)

// SchemaKind represents the type of schema
type SchemaKind string

const (
	SchemaKindJSON     SchemaKind = "jsonschema"
	SchemaKindExample  SchemaKind = "jsonexample"
	SchemaKindRef      SchemaKind = "ref"
)

// FlowStatus represents flow status
type FlowStatus string

const (
	FlowStatusRunning     FlowStatus = "RUNNING"
	FlowStatusSuspended   FlowStatus = "SUSPENDED"
	FlowStatusWaitingInput FlowStatus = "WAITING_INPUT"
	FlowStatusCompleted   FlowStatus = "COMPLETED"
	FlowStatusCancelled   FlowStatus = "CANCELLED"
	FlowStatusFailed      FlowStatus = "FAILED"
)

// EntityKind represents entity type
type EntityKind string

const (
	EntityKindUser  EntityKind = "user"
	EntityKindGroup EntityKind = "group"
	EntityKindRole  EntityKind = "role"
	EntityKindBot   EntityKind = "bot"
)

// Entity represents a routable target
type Entity struct {
	ID        string                 `json:"id"`
	Kind      EntityKind             `json:"kind"`
	Handle    string                 `json:"handle,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
	CreatedAt string                 `json:"createdAt,omitempty"`
}

// Request represents a data-entry request
type Request struct {
	ID            string                 `json:"id"`
	CreatedBy     string                 `json:"createdBy"`
	EntityID      string                 `json:"entityId"`
	Status        Status                 `json:"status"`
	SchemaKind    SchemaKind             `json:"schemaKind"`
	SchemaPayload map[string]interface{} `json:"schemaPayload"`
	UIHints       map[string]interface{} `json:"uiHints,omitempty"`
	Prefill       map[string]interface{} `json:"prefill,omitempty"`
	ExpiresAt     *string                `json:"expiresAt,omitempty"`
	DeadlineAt    *string                `json:"deadlineAt,omitempty"`
	AttentionAt   *string                `json:"attentionAt,omitempty"`
	CallbackURL   *string                `json:"callbackUrl,omitempty"`
	FilesPolicy   map[string]interface{} `json:"filesPolicy,omitempty"`
	FlowID        *string                `json:"flowId,omitempty"`
	CreatedAt     string                 `json:"createdAt,omitempty"`
	UpdatedAt     string                 `json:"updatedAt,omitempty"`
}

// Response represents a response to a request
type Response struct {
	ID          string                 `json:"id"`
	RequestID   string                 `json:"requestId"`
	AnsweredBy  string                 `json:"answeredBy"`
	Payload     map[string]interface{} `json:"payload"`
	Files       []map[string]interface{} `json:"files,omitempty"`
	AnsweredAt string                 `json:"answeredAt,omitempty"`
}

// Flow represents a durable workflow
type Flow struct {
	ID          string                 `json:"id"`
	Kind        string                 `json:"kind"`
	OwnerEntity string                 `json:"ownerEntity"`
	Status      FlowStatus             `json:"status"`
	Cursor      map[string]interface{} `json:"cursor"`
	LastEventID *string                `json:"lastEventId,omitempty"`
	CreatedAt   string                 `json:"createdAt,omitempty"`
	UpdatedAt   string                 `json:"updatedAt,omitempty"`
}

