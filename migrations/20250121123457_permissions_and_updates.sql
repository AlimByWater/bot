-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS permissions (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    private_generation BOOLEAN NOT NULL DEFAULT FALSE,
    use_by_channel_name BOOLEAN NOT NULL DEFAULT FALSE,
    vip BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE users
    ADD COLUMN balance INT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS updates_log (
    bot_id BIGINT NOT NULL REFERENCES bots(id),
    update JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS updates_log;
DROP TABLE IF EXISTS permissions;
ALTER TABLE users
    DROP COLUMN balance;
-- +goose StatementEnd
