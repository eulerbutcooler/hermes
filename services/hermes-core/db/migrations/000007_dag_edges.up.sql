ALTER TABLE relay_actions ADD COLUMN IF NOT EXISTS node_id TEXT NOT NULL;

ALTER TABLE relay_actions
ADD CONSTRAINT uq_relay_actions_node_id UNIQUE(relay_id,node_id);

CREATE TABLE IF NOT EXISTS relay_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relay_id UUID NOT NULL REFERENCES relays(id) ON DELETE CASCADE,
    parent_node_id TEXT NOT NULL,
    child_node_id  TEXT NOT NULL,
    condition    JSONB, 
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(relay_id, parent_node_id, child_node_id),
    FOREIGN KEY (relay_id, parent_node_id) REFERENCES relay_actions(relay_id, node_id) ON DELETE CASCADE,
    FOREIGN KEY (relay_id, child_node_id)  REFERENCES relay_actions(relay_id, node_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_relay_edges_relay ON relay_edges(relay_id);
CREATE INDEX IF NOT EXISTS idx_relay_edges_child ON relay_edges(relay_id, child_node_id);

ALTER TABLE execution_steps ADD COLUMN IF NOT EXISTS node_id TEXT NOT NULL;

ALTER TABLE relay_actions DROP COLUMN IF EXISTS order_index CASCADE;
ALTER TABLE execution_steps DROP COLUMN IF EXISTS order_index CASCADE;

CREATE INDEX IF NOT EXISTS idx_execution_steps_node_id ON execution_steps(execution_id, node_id);