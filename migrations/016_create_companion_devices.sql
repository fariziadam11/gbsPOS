CREATE TABLE IF NOT EXISTS companion_devices (
    device_id VARCHAR(100) PRIMARY KEY,
    device_name VARCHAR(150),
    sdk_version VARCHAR(50),
    capabilities TEXT,
    last_seen_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
