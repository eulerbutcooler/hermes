-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Relays table (webhook workflows)
CREATE TABLE IF NOT EXISTS relays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Relay actions (steps in a workflow)
CREATE TABLE IF NOT EXISTS relay_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relay_id UUID NOT NULL REFERENCES relays(id) ON DELETE CASCADE,
    action_type TEXT NOT NULL,
    config JSONB NOT NULL,
    order_index INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(relay_id, order_index)
);

-- Execution logs (audit trail)
CREATE TABLE IF NOT EXISTS execution_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relay_id UUID NOT NULL REFERENCES relays(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    payload JSONB,
    error_message TEXT,
    executed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_relays_user_id ON relays(user_id);
CREATE INDEX IF NOT EXISTS idx_relays_webhook_path ON relays(webhook_path);
CREATE INDEX IF NOT EXISTS idx_relay_actions_relay_id ON relay_actions(relay_id);
CREATE INDEX IF NOT EXISTS idx_execution_logs_relay_id ON execution_logs(relay_id);
CREATE INDEX IF NOT EXISTS idx_execution_logs_executed_at ON execution_logs(executed_at DESC);

-- Insert test data
INSERT INTO users (id, username, email) VALUES
    ('000-1', 'testuser', 'test@hermes.dev')
ON CONFLICT (email) DO NOTHING;

INSERT INTO relays (id, user_id, name, description, webhook_path) VALUES
    (
        '000-2',
        '000-1',
        'Test Discord Relay',
        'Test webhook that sends to Discord',
        '/hooks/test-relay'
    )
ON CONFLICT (webhook_path) DO NOTHING;

INSERT INTO relay_actions (relay_id, action_type, config, order_index) VALUES
    (
        '000-2',
        'debug_log',
        '{"message": "Webhook received"}',
        1
    ),
    (
        '000-2',
        'discord_send',
        '{"webhook_url": "https://discord.com/api/webhooks/YOUR_ID_HERE", "message": "Test from Hermes"}',
        2
    )
ON CONFLICT (relay_id, order_index) DO NOTHING;
