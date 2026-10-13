CREATE UNIQUE INDEX IF NOT EXISTS idx_qris_active_btn_order
	ON pos_qris_transactions (order_id)
	WHERE provider = 'BTN_SNAP'
	  AND status IN ('PENDING', 'AWAITING_CONFIRMATION', 'UNKNOWN', 'PAID')
	  AND order_id <> '';
