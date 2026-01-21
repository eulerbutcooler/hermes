CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS relays(
    id TEXT PRIMARY KEY,
    user_id INT REFERENCES users(id),
    name TEXT NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS relay_actions(
    id SERIAL PRIMARY KEY,
    relay_id TEXT REFERENCES relays(id) ON DELETE CASCADE,
    action_type TEXT NOT NULL,
    config JSONB NOT NULL,
    sort_order INT DEFAULT 1
)

CREATE TABLE IF NOT EXISTS execution_logs(
    id SERIAL PRIMARY KEY,
    relay_id TEXT REFERENCES relays(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    details TEXT,
    executed_at TIMESTAMP DEFAULT NOW())

INSERT INTO users(username, email)
VALUES ('testuser', 'test@gmail.com`)
ON CONFLICT (email) DO NOTHING;
