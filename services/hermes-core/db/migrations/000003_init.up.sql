ALTER TABLE execution_logs ADD COLUMN IF NOT EXISTS event_id TEXT;
CREATE INDEX IF NOT EXISTS idx_execution_logs_events_id ON execution_logs(event_id)
