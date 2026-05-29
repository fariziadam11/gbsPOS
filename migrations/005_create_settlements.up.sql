CREATE TABLE settlements (
    id          VARCHAR(64)    PRIMARY KEY,
    timestamp   BIGINT         NOT NULL,
    batch_count INT            NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL,
    card_total  DECIMAL(12,2)  NOT NULL,
    qris_total  DECIMAL(12,2)  NOT NULL,
    cash_total  DECIMAL(12,2)  NOT NULL,
    status      VARCHAR(20)    NOT NULL CHECK (status IN ('SUCCESS', 'FAILED')),
    store_type  VARCHAR(20),
    terminal_id VARCHAR(32),
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_settlements_timestamp ON settlements(timestamp DESC);
