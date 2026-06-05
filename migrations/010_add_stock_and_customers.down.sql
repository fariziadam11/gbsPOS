-- Revert: remove stock and customer features
DROP INDEX IF EXISTS idx_orders_customer_id;
DROP INDEX IF EXISTS idx_stock_movements_product_id;
DROP INDEX IF EXISTS idx_customers_phone;

ALTER TABLE orders DROP COLUMN IF EXISTS loyalty_points_earned;
ALTER TABLE orders DROP COLUMN IF EXISTS customer_id;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS stock_movements;
ALTER TABLE products DROP COLUMN IF EXISTS low_stock_threshold;
ALTER TABLE products DROP COLUMN IF EXISTS stock_quantity;
