-- 012_add_variants_discounts.up.sql
-- Multi-dimensional product variants + order discounts

CREATE TABLE IF NOT EXISTS product_variants (
    id                SERIAL PRIMARY KEY,
    product_id        INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku               VARCHAR(100),
    name              VARCHAR(255) NOT NULL,
    attributes        JSONB NOT NULL DEFAULT '{}',
    price             DECIMAL(12,2) DEFAULT NULL,
    stock_quantity    INTEGER NOT NULL DEFAULT 0,
    low_stock_threshold INTEGER DEFAULT NULL,
    is_active         BOOLEAN NOT NULL DEFAULT true,
    sort_order        INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX IF NOT EXISTS idx_product_variants_sku ON product_variants(sku);

ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS variant_id INTEGER DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS variant_name VARCHAR(255) DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS sku VARCHAR(100) DEFAULT NULL;

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS discount_type VARCHAR(20) DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS discount_value DECIMAL(12,2) DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(12,2) DEFAULT NULL;
