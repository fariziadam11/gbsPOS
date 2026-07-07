CREATE TABLE IF NOT EXISTS pos_hold_sessions (
    id UUID PRIMARY KEY,

    store_type VARCHAR(20) NOT NULL CHECK (store_type IN ('RETAIL', 'FNB', 'OUTFIT')),
    terminal_id VARCHAR(50) NOT NULL,

    payload JSONB NOT NULL,
    total DECIMAL(12,2) NOT NULL DEFAULT 0,

    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'RESUMED', 'EXPIRED')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pos_hold_terminal ON pos_hold_sessions(terminal_id);
CREATE INDEX IF NOT EXISTS idx_pos_hold_status ON pos_hold_sessions(status);
CREATE INDEX IF NOT EXISTS idx_pos_hold_store_type ON pos_hold_sessions(store_type);