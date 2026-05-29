CREATE TABLE ads (
    id             SERIAL PRIMARY KEY,
    name           VARCHAR(200)   NOT NULL,
    filename       VARCHAR(255)   NOT NULL,
    storage_path   VARCHAR(500)   NOT NULL,
    file_size      BIGINT         NOT NULL,
    mime_type      VARCHAR(50)    NOT NULL,
    duration_seconds INT,
    store_types    VARCHAR(200)   NOT NULL,
    playlist_order INT            NOT NULL DEFAULT 0,
    is_active      BOOLEAN        NOT NULL DEFAULT TRUE,
    start_date     DATE,
    end_date       DATE,
    start_time     TIME,
    end_time       TIME,
    created_by     INT            NOT NULL REFERENCES users(id),
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ads_is_active ON ads(is_active);
CREATE INDEX idx_ads_dates ON ads(start_date, end_date);
