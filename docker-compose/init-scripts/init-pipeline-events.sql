-- Phase 4 Data Fabric: pipeline events table for NiFi and Kafka event logging.
-- Run once against devsecops-postgres (e.g. when Postgres is first initialized via
-- /docker-entrypoint-initdb.d, or manually: psql -U qminiwasm -d qminiwasm_db -f init-pipeline-events.sql).
-- Requires pgvector extension (pgvector/pgvector:pg16 image provides it).

CREATE EXTENSION IF NOT EXISTS vector;

CREATE SCHEMA IF NOT EXISTS pipeline;

CREATE TABLE IF NOT EXISTS pipeline.pipeline_events (
  id            BIGSERIAL PRIMARY KEY,
  event_time    TIMESTAMPTZ NOT NULL DEFAULT now(),
  source        VARCHAR(64) NOT NULL,   -- 'nifi' | 'kafka' | 'n8n'
  topic         VARCHAR(255),          -- Kafka topic or NiFi flow name
  event_type    VARCHAR(128),          -- e.g. flow_file_received, vuln_scan_completed
  payload       JSONB,                 -- event body
  embedding     vector(384),           -- optional: pgvector for similarity search (e.g. 384-dim)
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pipeline_events_event_time ON pipeline.pipeline_events (event_time);
CREATE INDEX IF NOT EXISTS idx_pipeline_events_source ON pipeline.pipeline_events (source);
CREATE INDEX IF NOT EXISTS idx_pipeline_events_topic ON pipeline.pipeline_events (topic);
CREATE INDEX IF NOT EXISTS idx_pipeline_events_event_type ON pipeline.pipeline_events (event_type);
CREATE INDEX IF NOT EXISTS idx_pipeline_events_payload_gin ON pipeline.pipeline_events USING gin (payload);

-- Optional: index for vector similarity (use when querying by embedding).
-- CREATE INDEX idx_pipeline_events_embedding ON pipeline.pipeline_events USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

COMMENT ON TABLE pipeline.pipeline_events IS 'Phase 4: Central log of pipeline events from NiFi and Kafka for observability and pgvector search.';
