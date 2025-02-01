-- +goose Up
-- +goose StatementBegin
-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id bigint UNIQUE ,
    telegram_username VARCHAR(255),
    firstname VARCHAR(255),
    date_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы песен
CREATE TABLE IF NOT EXISTS songs (
                                             id SERIAL PRIMARY KEY,
                                             url VARCHAR(255) UNIQUE NOT NULL CHECK (url <> ''),
                                             artist_name VARCHAR(255) NOT NULL CHECK (artist_name <> ''),
                                             title VARCHAR(255),
                                             cover_link VARCHAR(255),
                                             cover_telegram_file_id VARCHAR(255),
                                             song_telegram_message_chat_id bigint,
                                             song_telegram_message_id bigint,
                                             tags TEXT[],
                                             date_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для записи загрузок песен пользователями
CREATE TABLE IF NOT EXISTS user_song_downloads (
                                                           id SERIAL PRIMARY KEY,
                                                           user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                                           song_id INT NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
                                                           download_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для записи воспроизведения песен
CREATE TABLE IF NOT EXISTS song_plays (
                                                  id SERIAL PRIMARY KEY,
                                                  song_id INT NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
                                                  stream VARCHAR(255) NOT NULL DEFAULT 'main',
                                                  play_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для веб-приложения событий
CREATE TABLE IF NOT EXISTS web_app_events (
                                                      id SERIAL PRIMARY KEY,
                                                      event_type VARCHAR(50) NOT NULL,
                                                      telegram_id BIGINT NOT NULL REFERENCES users(telegram_id),
                                                      payload JSONB,
                                                      session_id VARCHAR(255) NOT NULL,
                                                      stream VARCHAR(255) NOT NULL DEFAULT 'main',
                                                      timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для хранения данных о продолжительности сессии
CREATE TABLE IF NOT EXISTS user_session_durations (
                                                              id SERIAL PRIMARY KEY,
                                                              telegram_id BIGINT NOT NULL REFERENCES users(telegram_id),
                                                              start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                                              end_time TIMESTAMP,
                                                              duration_in_seconds BIGINT
);

-- Создание таблицы для хранения прослушиваний песен пользователями
CREATE TABLE IF NOT EXISTS user_to_song_history (
                                                            telegram_id BIGINT NOT NULL REFERENCES users(telegram_id),
                                                            song_id int NOT NULL REFERENCES songs(id),
                                                            song_plays_id int NOT NULL REFERENCES song_plays(id),
                                                            stream VARCHAR(255) NOT NULL DEFAULT 'main',
                                                            timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS songs_downloads (
                                                       song_id INTEGER NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
                                                       user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                                       source VARCHAR(255) NOT NULL DEFAULT '',
                                                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS streams (
                                               slug VARCHAR(255) NOT NULL,
                                               name VARCHAR(255) NOT NULL,
                                               link VARCHAR(255) NOT NULL,
                                               logo_link VARCHAR(255) NOT NULL,
                                               icon_link VARCHAR(255) NOT NULL,
                                               on_click_link VARCHAR(255) NOT NULL,
                                               created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                               updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                               enabled BOOLEAN DEFAULT TRUE,
                                               PRIMARY KEY (slug)
);


-- Создание индексов для быстрого поиска и уникальности
CREATE INDEX IF NOT EXISTS idx_users_username ON users(telegram_username);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_songs_title ON songs(title);
CREATE INDEX IF NOT EXISTS idx_songs_title ON songs(artist_name);
CREATE INDEX IF NOT EXISTS idx_songs_title ON songs(url);
CREATE INDEX IF NOT EXISTS idx_user_song_downloads_user_id ON user_song_downloads(user_id);
CREATE INDEX IF NOT EXISTS idx_user_song_downloads_song_id ON user_song_downloads(song_id);
CREATE INDEX IF NOT EXISTS idx_song_plays_song_id ON song_plays(song_id);
CREATE INDEX IF NOT EXISTS idx_web_app_events_telegram_id ON web_app_events(telegram_id);
CREATE INDEX IF NOT EXISTS idx_web_app_events_session_id ON web_app_events(session_id);
CREATE INDEX IF NOT EXISTS idx_web_app_events_event_type ON web_app_events(event_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS songs_downloads;

DROP TABLE IF EXISTS user_to_song_history;

DROP TABLE IF EXISTS user_session_durations;

DROP TABLE IF EXISTS web_app_events;

-- Удаление таблицы воспроизведения песен
DROP TABLE IF EXISTS song_plays;

-- Удаление таблицы загрузок песен пользователями
DROP TABLE IF EXISTS user_song_downloads;

-- Удаление таблицы песен
DROP TABLE IF EXISTS songs;

-- Удаление таблицы пользователей
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
