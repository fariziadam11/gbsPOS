-- Migration: Alter qris_transaction status column to accommodate AWAITING_CONFIRMATION
-- Run this SQL to update existing database

-- Check if column exists and alter it
ALTER TABLE pos_qris_transactions
ALTER COLUMN status TYPE VARCHAR(30);

-- Verify the change
-- SELECT column_name, data_type, character_maximum_length
-- FROM information_schema.columns
-- WHERE table_name = 'pos_qris_transactions' AND column_name = 'status';
