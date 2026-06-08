-- 012_add_variants_discounts.down.sql
ALTER TABLE orders
    DROP COLUMN IF EXISTS discount_type,
    DROP COLUMN IF EXISTS discount_value,
    DROP COLUMN IF EXISTS discount_amount;

ALTER TABLE order_items
    DROP COLUMN IF EXISTS variant_id,
    DROP COLUMN IF EXISTS variant_name,
    DROP COLUMN IF EXISTS sku;

DROP TABLE IF EXISTS product_variants;
