-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS bot_updates
(
    bot_id      Int64,                   -- Идентификатор бота (обязательно)
    update_time DateTime64(3) DEFAULT now(),   -- Время получения обновления
    payload     String                    -- Полезная нагрузка (например, JSON)
)
    ENGINE = MergeTree
PARTITION BY toYYYYMMDD(update_time)
ORDER BY (bot_id, update_time);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bot_updates;
-- +goose StatementEnd
