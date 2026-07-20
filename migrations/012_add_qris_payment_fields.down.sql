-- Rollback: Remove QRIS payment gateway fields from orders table

ALTER TABLE orders DROP COLUMN IF EXISTS qris_payment_id;
ALTER TABLE orders DROP COLUMN IF EXISTS qris_status;
ALTER TABLE orders DROP COLUMN IF EXISTS qris_link_url;
ALTER TABLE orders DROP COLUMN IF EXISTS qris_expires_at;
ALTER TABLE orders DROP COLUMN IF EXISTS qris_fee;
ALTER TABLE orders DROP COLUMN IF EXISTS qris_net_amount;
ALTER TABLE orders DROP COLUMN IF EXISTS qris_completed_at;

DROP INDEX IF EXISTS idx_orders_qris_payment_id;
DROP INDEX IF EXISTS idx_orders_qris_status;
