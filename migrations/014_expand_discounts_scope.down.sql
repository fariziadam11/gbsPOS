DROP INDEX IF EXISTS idx_discounts_voucher_code;
DROP INDEX IF EXISTS idx_discounts_scope;

ALTER TABLE discounts
    DROP CONSTRAINT IF EXISTS discounts_min_transaction_check,
    DROP CONSTRAINT IF EXISTS discounts_scope_check;

DELETE FROM discounts
WHERE product_id IS NULL;

ALTER TABLE discounts
    ALTER COLUMN product_id SET NOT NULL;

ALTER TABLE discounts
    DROP COLUMN IF EXISTS min_transaction,
    DROP COLUMN IF EXISTS voucher_code,
    DROP COLUMN IF EXISTS scope;
