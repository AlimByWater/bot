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

INSERT INTO bots (id, name, token, purpose, test, enabled)
VALUES (-1007894673045, 'optimus_polygon_bot', '7894673045:AAHgosEAHjdW78q44bPTSuwVqSZl8SEN0-w', 'emoji-gen-vip', true, true);

-- Создание таблицы для хранения информации о пользователях ботов
CREATE TABLE IF NOT EXISTS user_to_bots
(
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bot_id BIGINT NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, bot_id)
);


CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    bot_id BIGINT NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price INT NOT NULL CHECK (price >= 0),  -- Стоимость в токенах
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (bot_id, name)
);

CREATE TABLE IF NOT EXISTS emoji_packs
(
    id SERIAL PRIMARY KEY NOT NULL,
    creator_telegram_id BIGINT NOT NULL REFERENCES users(telegram_id),
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
DROP TABLE IF EXISTS access_hashes;
DROP TABLE IF EXISTS emoji_packs;
DROP TABLE IF EXISTS user_to_bots;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS bots;
-- +goose StatementEnd
