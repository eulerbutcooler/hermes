-- Event deduplication table
CREATE TABLE IF NOT EXISTS processed_events (
    relay_id   UUID NOT NULL REFERENCES relays(id) ON DELETE CASCADE,
    event_id   TEXT NOT NULL,
    received_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (relay_id, event_id)
);

CREATE INDEX IF NOT EXISTS idx_processed_events_received_at ON processed_events(received_at);

DROP INDEX IF EXISTS idx_processed_events_received_at;
DROP TABLE IF EXISTS processed_events;
