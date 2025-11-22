-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    user_id TEXT NOT NULL,
    start_time bigint NOT NULL,
    end_time bigint NOT NULL,
    notify_before INTEGER
);
CREATE INDEX IF NOT EXISTS idx_events_user_id ON events(user_id);
CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time);

-- +goose Down
DROP INDEX IF EXISTS idx_events_start_time;
DROP INDEX IF EXISTS idx_events_user_id;
DROP TABLE IF EXISTS events; 