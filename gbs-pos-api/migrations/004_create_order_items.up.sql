CREATE TABLE order_items (
    id             SERIAL PRIMARY KEY,
    order_id       VARCHAR(32)    NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id     INT            NOT NULL,
    product_name   VARCHAR(200)   NOT NULL,
    product_price  DECIMAL(12,2)  NOT NULL,
    qty            INT            NOT NULL CHECK (qty > 0),
    subtotal       DECIMAL(12,2)  NOT NULL
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
