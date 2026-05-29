CREATE TABLE products (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(200)   NOT NULL,
    price       DECIMAL(12,2)  NOT NULL CHECK (price >= 0),
    category    VARCHAR(100)   NOT NULL,
    image_url   VARCHAR(500),
    store_type  VARCHAR(20)    NOT NULL CHECK (store_type IN ('RETAIL', 'FNB', 'OUTFIT')),
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_store_type ON products(store_type);
CREATE INDEX idx_products_category ON products(category);
