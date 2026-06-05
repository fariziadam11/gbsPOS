-- Add stock tracking to products
ALTER TABLE products ADD COLUMN stock_quantity INTEGER NOT NULL DEFAULT 0;
ALTER TABLE products ADD COLUMN low_stock_threshold INTEGER NOT NULL DEFAULT 10;

-- Stock movements audit trail
CREATE TABLE stock_movements (
    id SERIAL PRIMARY KEY,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('IN', 'OUT', 'ADJUSTMENT')),
    quantity INTEGER NOT NULL,
    reason VARCHAR(255),
    reference_id VARCHAR(100),
    created_by VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Customers table
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    phone VARCHAR(50) UNIQUE,
    email VARCHAR(255),
    address TEXT,
    loyalty_points INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Link customers to orders
ALTER TABLE orders ADD COLUMN customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL;
ALTER TABLE orders ADD COLUMN loyalty_points_earned INTEGER NOT NULL DEFAULT 0;

-- Index for faster customer lookups
CREATE INDEX idx_customers_phone ON customers(phone);
CREATE INDEX idx_stock_movements_product_id ON stock_movements(product_id);
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
