CREATE TABLE orders (
    id              VARCHAR(32)    PRIMARY KEY,
    subtotal        DECIMAL(12,2)  NOT NULL,
    tax             DECIMAL(12,2)  NOT NULL,
    total           DECIMAL(12,2)  NOT NULL,
    payment_method  VARCHAR(20)    NOT NULL CHECK (payment_method IN ('CASH', 'CARD', 'QRIS')),
    cash_received   DECIMAL(12,2),
    change_amount   DECIMAL(12,2),
    timestamp       BIGINT         NOT NULL,
    is_voided       BOOLEAN        NOT NULL DEFAULT FALSE,
    is_settled      BOOLEAN        NOT NULL DEFAULT FALSE,
    transaction_id  VARCHAR(100),
    approval_code   VARCHAR(50),
    entry_mode      VARCHAR(20),
    masked_account  VARCHAR(50),
    acq_mid         VARCHAR(50),
    acq_tid         VARCHAR(50),
    pos_message_id  VARCHAR(100),
    bank_name       VARCHAR(50),
    store_type      VARCHAR(20),
    terminal_id     VARCHAR(32),
    void_reason     VARCHAR(255),
    voided_by       VARCHAR(50),
    voided_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_timestamp ON orders(timestamp DESC);
CREATE INDEX idx_orders_is_settled ON orders(is_settled);
CREATE INDEX idx_orders_is_voided ON orders(is_voided);
CREATE INDEX idx_orders_store_type ON orders(store_type);
CREATE INDEX idx_orders_terminal_id ON orders(terminal_id);
CREATE INDEX idx_orders_transaction_id ON orders(transaction_id);
