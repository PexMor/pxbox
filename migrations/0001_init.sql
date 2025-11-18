-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Entities table
CREATE TABLE entities (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  kind TEXT NOT NULL CHECK (kind IN ('user','group','role','bot')),
  handle TEXT UNIQUE,
  meta JSONB NOT NULL DEFAULT '{}'::JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_entities_handle ON entities(handle);

-- Flows table (created before requests to allow foreign key reference)
CREATE TABLE flows (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  kind TEXT NOT NULL,
  owner_entity UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK (status IN ('RUNNING','SUSPENDED','WAITING_INPUT','COMPLETED','CANCELLED','FAILED')) DEFAULT 'RUNNING',
  cursor JSONB NOT NULL DEFAULT '{}'::JSONB,
  last_event_id TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_flows_owner_status ON flows(owner_entity, status);
CREATE INDEX idx_flows_status ON flows(status);

-- Requests (inquiries) table
CREATE TABLE requests (
  id TEXT PRIMARY KEY, -- ULID
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL, -- requestor client_id (issuer)
  entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK (status IN ('PENDING','CLAIMED','ANSWERED','CANCELLED','EXPIRED')) DEFAULT 'PENDING',
  schema_kind TEXT NOT NULL CHECK (schema_kind IN ('jsonschema','jsonexample','ref')),
  schema_payload JSONB NOT NULL,
  ui_hints JSONB NOT NULL DEFAULT '{}'::JSONB,
  prefill JSONB,
  expires_at TIMESTAMPTZ,
  deadline_at TIMESTAMPTZ,
  attention_at TIMESTAMPTZ,
  autocancel_grace INTERVAL,
  callback_url TEXT,
  callback_secret TEXT,
  files_policy JSONB,
  flow_id UUID REFERENCES flows(id) ON DELETE SET NULL,
  deleted_at TIMESTAMPTZ,
  read_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_requests_entity_status ON requests(entity_id, status);
CREATE INDEX idx_requests_flow_id ON requests(flow_id);
CREATE INDEX idx_requests_deadline_at ON requests(deadline_at);
CREATE INDEX idx_requests_attention_at ON requests(attention_at);
CREATE INDEX idx_requests_status ON requests(status);

-- Responses table
CREATE TABLE responses (
  id TEXT PRIMARY KEY, -- ULID
  request_id TEXT NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
  answered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  answered_by UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  payload JSONB NOT NULL,
  files JSONB NOT NULL DEFAULT '[]'::JSONB,
  signature_jws TEXT
);

CREATE INDEX idx_responses_request_id ON responses(request_id);
CREATE INDEX idx_responses_answered_by ON responses(answered_by);


-- Reminders table
CREATE TABLE reminders (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  request_id TEXT NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  remind_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reminders_remind_at ON reminders(remind_at);
CREATE INDEX idx_reminders_request_id ON reminders(request_id);

-- Webhooks table (for future use)
CREATE TABLE webhooks (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  owner TEXT NOT NULL,
  url TEXT NOT NULL,
  secret TEXT NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  events TEXT[] NOT NULL DEFAULT ARRAY['request.created','request.answered']
);

