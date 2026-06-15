ALTER TABLE discounts
    ADD COLUMN IF NOT EXISTS scope VARCHAR(20) NOT NULL DEFAULT 'PRODUCT',
    ADD COLUMN IF NOT EXISTS voucher_code VARCHAR(100),
    ADD COLUMN IF NOT EXISTS min_transaction DECIMAL(12,2) DEFAULT 0;

UPDATE discounts
SET scope = 'PRODUCT'
WHERE scope IS NULL OR scope = '';

ALTER TABLE discounts
    ALTER COLUMN product_id DROP NOT NULL;

ALTER TABLE discounts
    ADD CONSTRAINT discounts_scope_check
        CHECK (scope IN ('PRODUCT', 'TRANSACTION', 'VOUCHER')),
    ADD CONSTRAINT discounts_min_transaction_check
        CHECK (min_transaction IS NULL OR min_transaction >= 0);

CREATE INDEX IF NOT EXISTS idx_discounts_scope ON discounts(scope);
CREATE INDEX IF NOT EXISTS idx_discounts_voucher_code ON discounts(voucher_code);
