-- Add QRIS payment gateway fields to orders table
-- Migration: 012_add_qris_payment_fields

ALTER TABLE orders ADD COLUMN IF NOT EXISTS qris_payment_id VARCHAR(100);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS qris_status VARCHAR(20) DEFAULT 'pending';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS qris_link_url TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS qris_expires_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS qris_fee DECIMAL(12,2) DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS qris_net_amount DECIMAL(12,2) DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS qris_completed_at TIMESTAMP WITH TIME ZONE;

COMMENT ON COLUMN orders.qris_payment_id IS 'SumoPod payment ID';
COMMENT ON COLUMN orders.qris_status IS 'QRIS payment status: pending, completed, failed, expired';
COMMENT ON COLUMN orders.qris_link_url IS 'Payment link URL for QRIS scanning';
COMMENT ON COLUMN orders.qris_expires_at IS 'Payment link expiration time';
COMMENT ON COLUMN orders.qris_fee IS 'SumoPod transaction fee';
COMMENT ON COLUMN orders.qris_net_amount IS 'Amount after fee deduction';
COMMENT ON COLUMN orders.qris_completed_at IS 'Timestamp when payment was completed';

CREATE INDEX IF NOT EXISTS idx_orders_qris_payment_id ON orders(qris_payment_id);
CREATE INDEX IF NOT EXISTS idx_orders_qris_status ON orders(qris_status);
