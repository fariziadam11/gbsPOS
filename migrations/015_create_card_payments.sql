CREATE TABLE IF NOT EXISTS card_payments (
    id UUID PRIMARY KEY,
    order_id VARCHAR(50) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'WAITING_FOR_CARD',
    device_id VARCHAR(100),
    terminal_id VARCHAR(100),
    transaction_id VARCHAR(100),
    card_brand VARCHAR(20),
    masked_card VARCHAR(30),
    auth_code VARCHAR(30),
    failure_reason TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_card_payments_status ON card_payments(status);
CREATE INDEX IF NOT EXISTS idx_card_payments_device ON card_payments(device_id);
CREATE INDEX IF NOT EXISTS idx_card_payments_terminal ON card_payments(terminal_id);
CREATE INDEX IF NOT EXISTS idx_card_payments_expires ON card_payments(expires_at);
