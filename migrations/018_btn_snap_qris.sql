ALTER TABLE pos_qris_transactions ALTER COLUMN static_qris_string DROP NOT NULL;
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS partner_reference_no VARCHAR(64);
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS bank_reference_no VARCHAR(64);
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS bank_payment_reference VARCHAR(20);
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS bank_status VARCHAR(2);
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS bank_status_description VARCHAR(50);
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS merchant_id VARCHAR(64);
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS currency VARCHAR(3);
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS bank_paid_at TIMESTAMPTZ;
ALTER TABLE pos_qris_transactions ADD COLUMN IF NOT EXISTS checkout_snapshot JSONB;
CREATE UNIQUE INDEX IF NOT EXISTS idx_qris_partner_reference
	ON pos_qris_transactions (partner_reference_no)
	WHERE partner_reference_no <> '';

ALTER TABLE orders ADD COLUMN IF NOT EXISTS ppob_items JSONB;
