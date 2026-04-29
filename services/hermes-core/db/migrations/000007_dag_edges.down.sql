ALTER TABLE relay_actions ADD COLUMN IF NOT EXISTS order_index INT;
ALTER TABLE execution_steps ADD COLUMN IF NOT EXISTS order_index INT;

DROP TABLE IF EXISTS relay_edges;

ALTER TABLE relay_actions DROP CONSTRAINT IF EXISTS uq_relay_actions_node_id;
ALTER TABLE relay_actions DROP COLUMN IF EXISTS node_id;

DROP INDEX IF EXISTS idx_execution_steps_node_id;
ALTER TABLE execution_steps DROP COLUMN IF EXISTS node_id;

