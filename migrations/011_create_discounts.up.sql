CREATE TABLE IF NOT EXISTS discounts (
    id          SERIAL PRIMARY KEY,
    product_id  INTEGER         NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name        VARCHAR(200)    NOT NULL,
    type        VARCHAR(20)     NOT NULL CHECK (type IN ('PERCENTAGE', 'FIXED')),
    value       DECIMAL(12,2)   NOT NULL CHECK (value > 0),
    start_date  TIMESTAMPTZ     NOT NULL,
    end_date    TIMESTAMPTZ     NOT NULL,
    status      VARCHAR(20)     NOT NULL CHECK (status IN ('PENDING', 'ACTIVE', 'EXPIRED', 'STOPPED', 'CANCELLED')),
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    CONSTRAINT discounts_date_range_check CHECK (start_date <= end_date),
    CONSTRAINT discounts_percentage_value_check CHECK (type <> 'PERCENTAGE' OR value <= 100)
);

CREATE INDEX IF NOT EXISTS idx_discounts_product_id ON discounts(product_id);
CREATE INDEX IF NOT EXISTS idx_discounts_product_period ON discounts(product_id, start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_discounts_status ON discounts(status);
