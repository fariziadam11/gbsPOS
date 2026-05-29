CREATE TABLE ad_play_logs (
    id          SERIAL PRIMARY KEY,
    ad_id       INT            NOT NULL REFERENCES ads(id),
    terminal_id VARCHAR(32),
    store_type  VARCHAR(20)    NOT NULL,
    played_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ad_play_logs_ad_id ON ad_play_logs(ad_id);
CREATE INDEX idx_ad_play_logs_played_at ON ad_play_logs(played_at);
