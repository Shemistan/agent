CREATE TABLE IF NOT EXISTS health_calls (
    id SERIAL PRIMARY KEY,
    called_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_health_calls_called_at ON health_calls(called_at);
