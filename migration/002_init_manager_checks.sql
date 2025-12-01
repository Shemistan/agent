CREATE TABLE IF NOT EXISTS manager_checks (
    id SERIAL PRIMARY KEY,
    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    manager_url TEXT NOT NULL,
    status TEXT NOT NULL,
    http_status INT NULL,
    error_message TEXT NULL
);

CREATE INDEX IF NOT EXISTS idx_manager_checks_checked_at ON manager_checks(checked_at);
CREATE INDEX IF NOT EXISTS idx_manager_checks_status ON manager_checks(status);
