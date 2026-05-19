-- GBS POS & CMS Database Schema
-- Generated from dbdiagram.dbml for PostgreSQL 15+
-- Both POS API and CMS API share this same database (gbs_pos)

CREATE TYPE enum_role AS ENUM ('ADMIN', 'CASHIER');
CREATE TYPE enum_store_type AS ENUM ('RETAIL', 'FNB', 'OUTFIT');
CREATE TYPE enum_payment_method AS ENUM ('CASH', 'CARD', 'QRIS');
CREATE TYPE enum_status AS ENUM ('SUCCESS', 'FAILED');

-- Users (shared between POS and CMS)
CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    username      VARCHAR(50)  UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(100),
    role          enum_role    NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_users_username ON users(username);

-- Products
CREATE TABLE products (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(200)   NOT NULL,
    price       DECIMAL(12,2)  NOT NULL CHECK (price >= 0),
    category    VARCHAR(100)   NOT NULL,
    image_url   VARCHAR(500),
    store_type  enum_store_type NOT NULL,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_store_type ON products(store_type);
CREATE INDEX idx_products_category ON products(category);

-- Orders
CREATE TABLE orders (
    id              VARCHAR(32)    PRIMARY KEY,
    subtotal        DECIMAL(12,2)  NOT NULL,
    tax             DECIMAL(12,2)  NOT NULL,
    total           DECIMAL(12,2)  NOT NULL,
    payment_method  enum_payment_method NOT NULL,
    cash_received   DECIMAL(12,2),
    change_amount   DECIMAL(12,2),
    timestamp       BIGINT         NOT NULL,
    is_voided       BOOLEAN        NOT NULL DEFAULT FALSE,
    is_settled      BOOLEAN        NOT NULL DEFAULT FALSE,
    transaction_id  VARCHAR(100),
    approval_code   VARCHAR(50),
    entry_mode      VARCHAR(20),
    masked_account  VARCHAR(50),
    acq_mid         VARCHAR(50),
    acq_tid         VARCHAR(50),
    pos_message_id  VARCHAR(100),
    bank_name       VARCHAR(50),
    store_type      enum_store_type,
    terminal_id     VARCHAR(32),
    void_reason     VARCHAR(255),
    voided_by       VARCHAR(50),
    voided_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_timestamp ON orders(timestamp DESC);
CREATE INDEX idx_orders_is_settled ON orders(is_settled);
CREATE INDEX idx_orders_is_voided ON orders(is_voided);
CREATE INDEX idx_orders_store_type ON orders(store_type);
CREATE INDEX idx_orders_terminal_id ON orders(terminal_id);
CREATE INDEX idx_orders_transaction_id ON orders(transaction_id);

-- Order Items
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

-- Settlements
CREATE TABLE settlements (
    id           VARCHAR(64)    PRIMARY KEY,
    timestamp    BIGINT         NOT NULL,
    batch_count  INT            NOT NULL,
    total_amount DECIMAL(12,2)  NOT NULL,
    card_total   DECIMAL(12,2)  NOT NULL,
    qris_total   DECIMAL(12,2)  NOT NULL,
    cash_total   DECIMAL(12,2)  NOT NULL,
    status       enum_status    NOT NULL,
    store_type   enum_store_type,
    terminal_id  VARCHAR(32),
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_settlements_timestamp ON settlements(timestamp DESC);

-- Ads (CMS)
CREATE TABLE ads (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(200)   NOT NULL,
    filename        VARCHAR(255)   NOT NULL,
    storage_path    VARCHAR(500)   NOT NULL,
    file_size       BIGINT         NOT NULL,
    mime_type       VARCHAR(50)    NOT NULL,
    duration_seconds INT,
    store_types     JSON           NOT NULL,
    playlist_order  INT            NOT NULL DEFAULT 0,
    is_active       BOOLEAN        NOT NULL DEFAULT TRUE,
    start_date      DATE,
    end_date        DATE,
    start_time      TIME,
    end_time        TIME,
    created_by      INT            NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ads_is_active ON ads(is_active);
CREATE INDEX idx_ads_dates ON ads(start_date, end_date);

-- Ad Play Logs (CMS analytics)
CREATE TABLE ad_play_logs (
    id          SERIAL PRIMARY KEY,
    ad_id       INT            NOT NULL REFERENCES ads(id),
    terminal_id VARCHAR(32),
    store_type  enum_store_type NOT NULL,
    played_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ad_play_logs_ad_id ON ad_play_logs(ad_id);
CREATE INDEX idx_ad_play_logs_played_at ON ad_play_logs(played_at);
