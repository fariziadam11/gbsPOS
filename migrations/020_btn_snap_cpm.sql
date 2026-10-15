ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS flow_type VARCHAR(8);
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS generate_time VARCHAR(25);
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS transaction_date_time VARCHAR(25);
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS scanner_device_id VARCHAR(20);
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS merchant_pan VARCHAR(19);
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS cancel_requested_at TIMESTAMPTZ;

UPDATE pos_qris_transactions SET flow_type = 'MPM'
WHERE provider = 'BTN_SNAP' AND (flow_type IS NULL OR flow_type = '');

CREATE INDEX IF NOT EXISTS idx_qris_order_flow_created
	ON pos_qris_transactions (order_id, flow_type, created_at DESC);

CREATE TABLE IF NOT EXISTS pos_qris_refunds (
	id VARCHAR(64) PRIMARY KEY,
	transaction_id VARCHAR(64) NOT NULL UNIQUE REFERENCES pos_qris_transactions(id),
	partner_refund_no VARCHAR(16) NOT NULL UNIQUE,
	refund_no VARCHAR(16) NOT NULL DEFAULT '',
	amount DECIMAL(12,2) NOT NULL,
	status VARCHAR(20) NOT NULL,
	reason VARCHAR(256) NOT NULL DEFAULT '',
	refunded_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pos_qris_refunds_status ON pos_qris_refunds (status);
