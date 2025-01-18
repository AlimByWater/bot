-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS bots
(
    id BIGINT PRIMARY KEY NOT NULL,
    name VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL,
    purpose VARCHAR(255) NOT NULL,
    test BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE
);

-- Создание таблицы для хранения информации о пользователях ботов
CREATE TABLE IF NOT EXISTS user_to_bots
(
    user_id INT NOT NULL REFERENCES users(id),
    bot_id BIGINT NOT NULL REFERENCES bots(id),
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS user_transactions
(
    id SERIAL PRIMARY KEY NOT NULL,
    user_id INT NOT NULL REFERENCES users(id),
    amount INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS emoji_packs
(
    id SERIAL PRIMARY KEY NOT NULL,
    creator_id INT NOT NULL REFERENCES users(id),
    bot_id BIGINT NOT NULL REFERENCES bots(id),
    pack_title VARCHAR(255) NOT NULL,
    pack_link VARCHAR(255) NOT NULL UNIQUE,
    telegram_file_id VARCHAR(255) NOT NULL,
    initial_command VARCHAR(255) NOT NULL,
    emoji_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS access_hashes
(
    chat_id VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL,
    hash BIGINT NOT NULL,
    peer_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO access_hashes (chat_id, username, hash, peer_id)
VALUES
('-1002400904088', '@drip_tech', 4253614615109204755, 2400904088), -- main forum @drip_tech
('-1002491830452', '@fullytestingpolygon', 1750568581171467725, 2491830452)
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS user_bots;
DROP TABLE IF EXISTS bots;
-- +goose StatementEnd
